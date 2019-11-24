package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"time"
	"bufio"
	"strings"
	"bytes"

	"github.com/geo-stanciu/go-utils/utils"
)

type configuration struct {
	DumpDir    string `json:"DumpDir"`
	Files2Keep int    `json:"Files2Keep"`
	User       string `json:"User"`
	Password   string `json:"Password"`
}

var (
	config          = configuration{}
	layout          = "20060102"
	currentDir      string
)

func init() {
	currentDir = filepath.Dir(os.Args[0])
}

func main() {
	var err error
	tNow := time.Now()
	t := tNow.UTC()
	sData := t.Format(layout)

	logFile, err := os.OpenFile(fmt.Sprintf("logs/backup_%s.txt", sData), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println(err)
		return
	}
	defer logFile.Close()

	mw := io.MultiWriter(os.Stdout, logFile)

	log.SetOutput(mw)

	cfgFile := fmt.Sprintf("%s/conf.json", currentDir)
	err = config.readFromFile(cfgFile)
	if err != nil {
		log.Println(err)
		return
	}

	dumpname := fmt.Sprintf("backup_%s.sql", sData)
	dumpFile := path.Join(config.DumpDir, dumpname)
	zipname := path.Join(config.DumpDir, fmt.Sprintf("backup_%s.zip", sData))

	log.Printf("start dump backup \"%s\"\n", dumpFile)

	outfile, err := os.Create(dumpFile)
	if err != nil {
		log.Println(err)
		return
	}

	cmd := exec.Command(
		"mysqldump",
		"-e",
		fmt.Sprintf("-u%s", config.User),
		fmt.Sprintf("-p%s", config.Password),
		"--single-transaction",
		"--flush-logs",
		"--all-databases",
	)

	cmd.Stdout = outfile

	err = cmd.Run()
	if err != nil {
		log.Println(err)
		return
	}

	outfile.Close()

	log.Printf("archive the dump file \"%s\"\n", zipname)

	zipfile, err := os.Create(zipname)
	if err != nil {
		log.Println(err)
		return
	}

	zip := utils.NewZipWriter(zipfile)
	zip.AddFile(dumpname, dumpFile)
	zip.Close()
	zipfile.Close()

	err = os.Remove(dumpFile)
	if err != nil {
		log.Println(err)
		return
	}
	
	sTimestamp := tNow.Format(utils.ISODateTime)
	sdt, err := getDate4Logs2BeRemoved("./backup.txt", sTimestamp);
	
	if len(sdt) > 0 {
		log.Printf("\n\nCleaning binary logs before \"%s\"\n", sdt)
		
		query := fmt.Sprintf(`
			PURGE BINARY LOGS BEFORE '%s'
		`,
			escapeString(sdt),
		)
		
		result, err := runQuery(query)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("%s\n", sdt, result)
	}

	directory := getAbsPath(config.DumpDir)

	if config.Files2Keep > 0 {
		log.Printf("\n\nCleaning old files from \"%s\"\n", directory)
		log.Printf("Will keep the last %d files.", config.Files2Keep)

		err = cleanDir(directory, "backup_*.zip")
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	log.Printf("\n\nend dump backup")
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

func runQuery(query string) (string, error) {
	var outb, errb bytes.Buffer

	cmd := exec.Command(
		"mysql",
		"-e", transform2SingleLine(query),
		fmt.Sprintf("-u%s", config.User),
		fmt.Sprintf("-p%s", config.Password),
		"-N",
		"-B",
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

func getAbsPath(dir string) string {
	directory := dir

	if len(directory) == 0 {
		directory = "./"
	}

	directory, err := filepath.Abs(directory)
	if err != nil {
		log.Fatal(err)
		return "./"
	}

	if directory[len(directory)-1:] == "/" || directory[len(directory)-1:] == "\\" {
		directory = directory[0 : len(directory)-1]
	}

	return directory
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

func getDate4Logs2BeRemoved(path string, sData string) (string, error) {
	bkFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return "", err
	}
	_, err = fmt.Fprintf(bkFile, "%s\n", sData)
	bkFile.Close()
	
	file, err := os.Open(path)
    if err != nil {
        return "", err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
	
	lines := make([]string, 0)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	
	toBeRemoved := len(lines) - config.Files2Keep - 1
	
	if toBeRemoved < 0 {
		return "", nil
	}
	
	return lines[toBeRemoved], nil
}

func cleanDir(directory, pattern string) error {
	files, err := filepath.Glob(directory + "/" + pattern)
	if err != nil {
		return err
	}

	sort.Slice(files, func(i, j int) bool {
		a := files[i]
		b := files[j]

		if len(a) >= 12 && len(b) >= 12 {
			sda := a[len(a)-12 : len(a)-4]
			sdb := b[len(b)-12 : len(b)-4]

			da, _ := time.Parse(layout, sda)
			db, _ := time.Parse(layout, sdb)

			return da.After(db)
		}

		return files[i] < files[j]
	})

	for i, f := range files {
		if i > config.Files2Keep-1 {
			log.Printf("deleting \"%s\"...\n", f)
			err = os.Remove(f)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *configuration) readFromFile(cfgFile string) error {
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
