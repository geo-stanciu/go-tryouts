package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

// Configuration - config struct
type Configuration struct {
	DbHost               string `json:"DbHost"`
	DbPort               string `json:"DbPort"`
	DbUser               string `json:"DbUser"`
	DbName               string `json:"DbName"`
	BackupDir            string `json:"BackupDir"`
	ArchiveDir           string `json:"ArchiveDir"`
	NumberOfBackups2Keep int    `json:"NumberOfBackups2Keep"`
}

var (
	config = Configuration{}
)

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

	/*
		On Windows:

		 You must edit C:\Users\geo\AppData\Roaming\postgresql\pgpass.conf on Windows
		 (1 row for each database !):

		 #hostname:port:database:username:password

		 On Linux:

		 su - postgres      //this will land in the home directory set for postgres user
		 vi .pgpass         //enter all users entries
		 chmod 0600 .pgpass // change the ownership to 0600 to avoid errors

		 For backup, PLEASE note the * used instead of a database name

		 #hostname:port:database:username:password
		 host:5432:*:postgres:password
	*/

	i := 0
	bkDirectory := ""
	bkLabel := ""

	for {
		bkDirectory = path.Join(config.BackupDir, fmt.Sprintf("%s_%02d", sData, i))
		bkLabel = fmt.Sprintf("BK %s base", fmt.Sprintf("%s %02d", sData, i))

		found, err := exists(bkDirectory)
		if err != nil {
			log.Println(err)
			return
		}

		if !found {
			break
		}

		i++
	}

	err = createBackupTables()
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("start backup with label \"%s\"\n", bkLabel)

	var outb, errb bytes.Buffer

	cmd := exec.Command(
		"pg_basebackup",
		"-D", bkDirectory,
		"-F", "t",
		"-r", "20M",
		"-R",
		"-X", "f",
		"-z",
		"-l", bkLabel,
		"-v",
		"-h", config.DbHost,
		"-p", config.DbPort,
		"-w",
		"-U", config.DbUser,
	)

	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	if err != nil {
		log.Println(err)
		return
	}

	sout := outb.String()
	serr := errb.String()
	logStd(sout, serr)

	archFile, err := getStartingArhiveLog(sout, serr)
	if err != nil {
		log.Println(err)
		return
	}

	err = logBackup(bkDirectory, archFile)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("\n\ncleanup:\n")

	err = keepOnlyNeededArchFiles(config.NumberOfBackups2Keep)
	if err != nil {
		log.Println(err)
		return
	}
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

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func logStd(sout, serr string) {

	so := strings.TrimSpace(sout)
	if len(so) > 0 {
		log.Println(so)
	}

	se := strings.TrimSpace(serr)
	if len(se) > 0 {
		log.Println(se)
	}
}

func transform2SingleLine(query string) string {
	q := query

	q = strings.Replace(q, "\r\n", " ", -1)
	q = strings.Replace(q, "\n", " ", -1)
	q = strings.Replace(q, "\r", " ", -1)

	return q
}

func escapeString(query string) string {
	q := query

	q = strings.Replace(q, "'", "''", -1)
	q = strings.Replace(q, "\\", "\\\\", -1)
	q = strings.Replace(q, "&", "' || chr(38) || '", -1)

	return q
}

func execQuery(query string) error {
	var outb, errb bytes.Buffer

	cmd := exec.Command(
		"psql",
		"-c", transform2SingleLine(query),
		"-d", config.DbName,
		"-h", config.DbHost,
		"-p", config.DbPort,
		"-w",
		"-U", config.DbUser,
	)

	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	logStd(outb.String(), errb.String())

	if err != nil {
		return err
	}

	return nil
}

func runQuery(query string) (string, error) {
	var outb, errb bytes.Buffer

	cmd := exec.Command(
		"psql",
		"-c", transform2SingleLine(query),
		"-d", config.DbName,
		"-h", config.DbHost,
		"-p", config.DbPort,
		"-t", // only the result
		"-w",
		"-U", config.DbUser,
	)

	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()

	sout := outb.String()
	serr := errb.String()

	if err != nil {
		return "", err
	}

	if len(sout) > 0 {
		return strings.TrimSpace(sout), nil
	}

	return strings.TrimSpace(serr), nil
}

func createBackupTables() error {
	query := `
		create table if not exists backup_log (
			backup_log_id   serial PRIMARY KEY,
			backup_time     timestamp not null default (now() at time zone 'UTC'),
			backup_dir      varchar(256) not null,
			arch_file       varchar(256) not null
		)
	`

	err := execQuery(query)
	if err != nil {
		return err
	}

	return nil
}

func getStartingArhiveLog(sout, serr string) (string, error) {
	look4 := "write-ahead log start point: "
	lookIn := sout

	idx := strings.Index(lookIn, look4)

	if idx < 0 {
		lookIn = serr
		idx = strings.Index(lookIn, look4)
	}

	if idx < 0 {
		return "", fmt.Errorf("write-ahead log start point not found")
	}

	slog := strings.TrimSpace(lookIn[idx+len(look4):])
	idx2 := strings.IndexAny(slog, " \t\r\n")

	if idx2 >= 0 {
		slog = strings.TrimSpace(slog[:idx2])
	}

	query := fmt.Sprintf(`
		SELECT file_name from pg_walfile_name_offset('%s')
	`, escapeString(slog))

	archFile, err := runQuery(query)
	if err != nil {
		return "", err
	}

	return archFile, nil
}

func logBackup(bkDirectory string, archFile string) error {

	query := fmt.Sprintf(`
		insert into backup_log (
			backup_dir,
			arch_file
		) values ('%s', '%s')
	`,
		escapeString(bkDirectory),
		escapeString(archFile),
	)

	err := execQuery(query)
	if err != nil {
		return err
	}

	return nil
}

func keepOnlyNeededArchFiles(nrBackups2Keep int) error {
	query := fmt.Sprintf(`
		WITH wals AS (
			select backup_log_id
			  from backup_log
			 order by backup_log_id desc
			 limit %d
		), min_wal AS (
			select min(backup_log_id) min_id from wals
		)
		select arch_file
		  from backup_log b
		  join min_wal m ON (b.backup_log_id = m.min_id)
	`, nrBackups2Keep)

	archFile, err := runQuery(query)
	if err != nil {
		return err
	}

	query = fmt.Sprintf(`
		WITH wals AS (
			select backup_log_id
			  from backup_log
			 order by backup_log_id desc
			 offset %d
		)
		select backup_dir
		  from backup_log b
		  join wals w ON (b.backup_log_id = w.backup_log_id)
	`, nrBackups2Keep)

	backupFiles, err := runQuery(query)
	if err != nil {
		return err
	}

	query = fmt.Sprintf(`
		WITH needed AS (
			select min(backup_log_id) min_id
			  from backup_log
		     where arch_file = '%s'
		), last_needed AS (
			select case when min_id is null then -1 else min_id end AS min_id
			  from needed
		)
		delete from backup_log b
		 where b.backup_log_id in (
			select backup_log_id
			  from backup_log, last_needed l
			 where backup_log_id < l.min_id
			 order by backup_log_id
		)
	`, escapeString(archFile))

	err = execQuery(query)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(strings.NewReader(backupFiles))
	for scanner.Scan() {
		bkFile := strings.TrimSpace(scanner.Text())

		log.Printf("Delete \"%s\"\n", bkFile)

		if _, err := os.Stat(bkFile); err == nil {
			err = os.RemoveAll(bkFile)
			if err != nil {
				return err
			}
		}
	}

	if len(archFile) > 0 {
		log.Printf("pg_archivecleanup -d %s %s\n", config.ArchiveDir, archFile)

		out, err := exec.Command("pg_archivecleanup", "-d", config.ArchiveDir, archFile).Output()
		if err != nil {
			return err
		}
		log.Println(string(out))
	}

	return nil
}
