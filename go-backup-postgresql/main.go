package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	_ "github.com/lib/pq"
)

type Configuration struct {
	DbURL                string `json:"DbURL"`
	PgDataDir            string `json:"PG-data-dir"`
	PgBackupDir          string `json:"PG-backup-dir"`
	PgArchiveDir         string `json:"PG-archive-dir"`
	NumberOfBackups2Keep int    `json:"NumberOfBackups2Keep"`
}

var (
	db     *sql.DB
	config = Configuration{}
)

func main() {
	t := time.Now()
	sData := t.Format("20060102")

	var err error

	logFile, err := os.OpenFile(fmt.Sprintf("logs/backup_%s.txt", sData), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		log.Fatal(err)
	}

	defer logFile.Close()

	mw := io.MultiWriter(os.Stdout, logFile)

	log.SetOutput(mw)

	cfgFile := "./conf.json"
	err = config.readFromFile(cfgFile)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	err = connect2Database(config.DbURL)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	defer db.Close()

	err = createBackupTables()

	if err != nil {
		log.Fatal(err)
	}

	bkFile, bkLabel, lastIndex, err := getBkFileName(sData)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("start backup with label \"%s\"\n", bkLabel)

	startBk, err := startBk(bkLabel)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("pg_start_backup: %s\n\n", startBk)

	out, err := exec.Command("jar", "cvf", bkFile, config.PgDataDir).Output()

	if err != nil {
		log.Fatal(err)
	}

	log.Println(string(out))

	archFile, err := finishBk()

	if err != nil {
		log.Fatal(err)
	}

	err = logBackup(bkFile, archFile, lastIndex)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("\n\ncleanup:\n")

	archFile2Keep, logID, err := getLastNeededArchFile(config.NumberOfBackups2Keep)

	err = deleteOldBackups(logID, archFile2Keep)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("stop backup \"%s\"\n\n\n", bkLabel)
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

func connect2Database(dbURL string) error {
	var err error

	db, err = sql.Open("postgres", dbURL)

	if err != nil {
		return errors.New("Can't connect to the database, go error " + fmt.Sprintf("%s", err))
	}

	err = db.Ping()

	if err != nil {
		return errors.New("Can't ping the database, go error " + fmt.Sprintf("%s", err))
	}

	return nil
}

func createBackupTables() error {
	t1 := `
		create table if not exists backup_log (
			backup_log_id   serial PRIMARY KEY,
			backup_time     timestamp DEFAULT statement_timestamp(),
			backup_file     varchar(256),
			arch_file       varchar(256),
			last_file_index varchar(8)
		)
	`

	_, err := db.Exec(t1)

	if err != nil {
		return err
	}

	return nil
}

func getBkFileName(sData string) (string, string, int, error) {
	var i int
	var bkFile string
	var bkLabel string

	query := `
		select CAST(last_file_index AS integer)
		  from backup_log
		 where backup_log_id = (
			 select max(backup_log_id)
			   from backup_log
			  where backup_time::date = to_date($1, 'yyyymmdd')
		 )
	`

	err := db.QueryRow(query, sData).Scan(&i)

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

func startBk(bkLabel string) (string, error) {
	var startBk string

	query := "SELECT pg_start_backup($1)::text"

	err := db.QueryRow(query, bkLabel).Scan(&startBk)

	switch {
	case err == sql.ErrNoRows:
		return "", err
	case err != nil:
		return "", err
	}

	return startBk, nil
}

func finishBk() (string, error) {
	var archFile2Keep string

	query := "SELECT file_name from pg_xlogfile_name_offset(pg_stop_backup())"

	err := db.QueryRow(query).Scan(&archFile2Keep)

	switch {
	case err == sql.ErrNoRows:
		return "", err
	case err != nil:
		return "", err
	}

	return archFile2Keep, nil
}

func logBackup(bkFile string, archFile string, lastFileIndex int) error {
	query := `
		insert into backup_log (
		    backup_file,
			arch_file,
			last_file_index
		) values ($1, $2, $3)
	`

	sIndex := fmt.Sprintf("%02d", lastFileIndex)

	_, err := db.Exec(query, bkFile, archFile, sIndex)

	if err != nil {
		return err
	}

	return nil
}

func getLastNeededArchFile(nrBackups2Keep int) (string, int, error) {
	var archFile string
	var logID int

	query := `
		select arch_file,
			   backup_log_id
		  from backup_log
		 order by backup_log_id desc
		 limit $1
	`

	rows, err := db.Query(query, nrBackups2Keep)

	if err != nil {
		return "", 0, err
	}

	defer rows.Close()

	i := 0

	for rows.Next() {
		i++

		err = rows.Scan(&archFile, &logID)

		if err != nil {
			return "", 0, err
		}
	}

	if err := rows.Err(); err != nil {
		return "", 0, err
	}

	if i < nrBackups2Keep {
		archFile = ""
		logID = -1
	}

	return archFile, logID, nil
}

func deleteOldBackups(logID int, archFile2Keep string) error {
	var bkFile string

	query := `
		select backup_file
		  from backup_log
		 where backup_log_id < $1
		 order by backup_log_id
	`

	rows, err := db.Query(query, logID)

	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&bkFile)

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

		if err != nil {
			return err
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	query = "delete from backup_log where backup_log_id < $1"

	_, err = db.Exec(query, logID)

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
