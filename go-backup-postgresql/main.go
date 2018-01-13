package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/geo-stanciu/go-utils/utils"
	_ "github.com/lib/pq"
)

type Configuration struct {
	DbType               string `json:"DbType"`
	DbURL                string `json:"DbURL"`
	PgDataDir            string `json:"PG-data-dir"`
	PgBackupDir          string `json:"PG-backup-dir"`
	PgArchiveDir         string `json:"PG-archive-dir"`
	NumberOfBackups2Keep int    `json:"NumberOfBackups2Keep"`
}

var (
	db      *sql.DB
	dbUtils *utils.DbUtils
	config  = Configuration{}
)

func init() {
	// init databaseutils
	dbUtils = new(utils.DbUtils)
}

func main() {
	t := time.Now().UTC()
	sData := t.Format("20060102")

	var err error

	logFile, err := os.OpenFile(fmt.Sprintf("logs/backup_%s.txt", sData), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println(err)
		return
	}
	defer logFile.Close()

	mw := io.MultiWriter(os.Stdout, logFile)

	log.SetOutput(mw)

	cfgFile := "./conf.json"
	err = config.readFromFile(cfgFile)
	if err != nil {
		log.Println(err)
		return
	}

	err = dbUtils.Connect2Database(&db, config.DbType, config.DbURL)
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		return
	}
	defer tx.Rollback()

	err = createBackupTables(tx)
	if err != nil {
		log.Println(err)
		return
	}

	bkFile, bkLabel, lastIndex, err := getBkFileName(tx, sData)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("start backup with label \"%s\"\n", bkLabel)

	startBk, err := startBk(tx, bkLabel)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("pg_start_backup: %s\n\n", startBk)

	out, err := exec.Command("jar", "cvf", bkFile, config.PgDataDir).Output()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(string(out))

	archFile, err := finishBk(tx)
	if err != nil {
		log.Println(err)
		return
	}

	err = logBackup(tx, bkFile, archFile, lastIndex)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("\n\ncleanup:\n")

	archFile2Keep, logID, err := getLastNeededArchFile(tx, config.NumberOfBackups2Keep)
	if err != nil {
		log.Println(err)
		return
	}

	err = deleteOldBackups(tx, logID, archFile2Keep)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("stop backup \"%s\"\n\n\n", bkLabel)

	tx.Commit()
}

func (c *Configuration) readFromFile(cfgFile string) error {
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		return err
	}

	file, err := os.Open(cfgFile)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&c)
	if err != nil {
		return err
	}

	return nil
}

func createBackupTables(tx *sql.Tx) error {
	t1 := `
		create table if not exists backup_log (
			backup_log_id   serial PRIMARY KEY,
			backup_time     timestamp not null,
			backup_file     varchar(256) not null,
			arch_file       varchar(256) not null,
			last_file_index varchar(8)   not null
		)
	`

	_, err := tx.Exec(t1)
	if err != nil {
		return err
	}

	return nil
}

func getBkFileName(tx *sql.Tx, sData string) (string, string, int, error) {
	var i int
	var bkFile string
	var bkLabel string

	pq := dbUtils.PQuery(`
		select CAST(last_file_index AS integer)
		  from backup_log
		 where backup_log_id = (
			 select max(backup_log_id)
			   from backup_log
			  where backup_time::date = to_date(?, 'yyyymmdd')
		 )
	`, sData)

	err := tx.QueryRow(pq.Query, pq.Args...).Scan(&i)

	switch {
	case err == sql.ErrNoRows:
		i = 0
	case err != nil:
		return "", "", 0, err
	}

	for {
		bkFile = fmt.Sprintf("%s/data_%s_%02d.zip", config.PgBackupDir, sData, i)
		bkLabel = fmt.Sprintf("BK %s %02d", sData, i)

		if _, err := os.Stat(bkFile); err == nil {
			i++
			continue
		} else {
			break
		}
	}

	return bkFile, bkLabel, i, nil
}

func startBk(tx *sql.Tx, bkLabel string) (string, error) {
	var startBk string

	pq := dbUtils.PQuery("SELECT pg_start_backup(?)::text", bkLabel)

	err := tx.QueryRow(pq.Query, pq.Args...).Scan(&startBk)

	switch {
	case err == sql.ErrNoRows:
		return "", err
	case err != nil:
		return "", err
	}

	return startBk, nil
}

func finishBk(tx *sql.Tx) (string, error) {
	var archFile2Keep string

	query := "SELECT file_name from pg_walfile_name_offset(pg_stop_backup())"

	err := tx.QueryRow(query).Scan(&archFile2Keep)

	switch {
	case err == sql.ErrNoRows:
		return "", err
	case err != nil:
		return "", err
	}

	return archFile2Keep, nil
}

func logBackup(tx *sql.Tx, bkFile string, archFile string, lastFileIndex int) error {
	dt := time.Now().UTC()
	sIndex := fmt.Sprintf("%02d", lastFileIndex)

	pq := dbUtils.PQuery(`
		insert into backup_log (
			backup_time,
			backup_file,
			arch_file,
			last_file_index
		) values (?, ?, ?, ?)
	`, dt,
		bkFile,
		archFile,
		sIndex)

	_, err := tx.Exec(pq.Query, pq.Args...)
	if err != nil {
		return err
	}

	return nil
}

func getLastNeededArchFile(tx *sql.Tx, nrBackups2Keep int) (string, int, error) {
	var archFile string
	var logID int
	var err error

	pq := dbUtils.PQuery(`
		select arch_file,
			   backup_log_id
		  from backup_log
		 order by backup_log_id desc
		 LIMIT ?
	`, nrBackups2Keep)

	i := 0

	err = dbUtils.ForEachRow(pq, func(row *sql.Rows) error {
		i++
		err = row.Scan(&archFile, &logID)
		return err
	})

	if err != nil {
		return "", 0, err
	}

	if i < nrBackups2Keep {
		archFile = ""
		logID = -1
	}

	return archFile, logID, nil
}

func deleteOldBackups(tx *sql.Tx, logID int, archFile2Keep string) error {
	var bkFile string
	var err error

	pq := dbUtils.PQuery(`
		select backup_file
		  from backup_log
		 where backup_log_id < ?
		 order by backup_log_id
	`, logID)

	err = dbUtils.ForEachRowTx(tx, pq, func(row *sql.Rows) error {
		err = row.Scan(&bkFile)
		if err != nil {
			return err
		}

		log.Printf("Delete \"%s\"\n", bkFile)

		if _, err := os.Stat(bkFile); err == nil {
			err = os.Remove(bkFile)
			if err != nil {
				return err
			}
		}

		return err
	})

	if err != nil {
		return err
	}

	pq = dbUtils.PQuery(`
		delete from backup_log where backup_log_id < ?
	`, logID)

	_, err = dbUtils.ExecTx(tx, pq)
	if err != nil {
		return err
	}

	if len(archFile2Keep) > 0 {
		log.Printf("pg_archivecleanup -d %s %s\n", config.PgArchiveDir, archFile2Keep)

		out, err := exec.Command("pg_archivecleanup", "-d", config.PgArchiveDir, archFile2Keep).Output()
		if err != nil {
			return err
		}
		log.Println(string(out))
	}

	return nil
}
