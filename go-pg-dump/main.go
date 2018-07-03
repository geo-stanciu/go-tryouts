package main

import (
	"bytes"
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
)

type configuration struct {
	DumpDir    string `json:"DumpDir"`
	Files2Keep int    `json:"Files2Keep"`
	DbName     string `json:"DbName"`
}

var (
	config = configuration{}
	layout = "20060102"
)

func main() {
	var err error
	t := time.Now().UTC()
	sData := t.Format(layout)

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

	dumpFile := path.Join(config.DumpDir, fmt.Sprintf("save_devel_%s.bak", sData))

	log.Printf("start dump backup \"%s\"\n", dumpFile)

	/*
		On Windows:

		 You must edit C:\Users\geo\AppData\Roaming\postgresql\pgpass.conf on Windows
		 (1 row for each database !):

		 #hostname:port:database:username:password

		 On Linux:

		 su - postgres      //this will land in the home directory set for postgres user
		 vi .pgpass         //enter all users entries
		 chmod 0600 .pgpass // change the ownership to 0600 to avoid errors

		 #hostname:port:database:username:password
	*/

	/*
		Restore with
		pg_restore -d devel -U postgres -Fc -C save_devel_yyyymmdd.bak
	*/

	var outb, errb bytes.Buffer

	cmd := exec.Command(
		"pg_dump",
		"-f", dumpFile,
		"--clean",
		"--format=c",
		"-v",
		config.DbName)

	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(outb.String())
	log.Println("Error:", errb.String())

	directory := getAbsPath(config.DumpDir)

	if config.Files2Keep > 0 {
		log.Printf("Cleaning old files from \"%s\"\n", directory)
		log.Printf("Will keep the last %d files.", config.Files2Keep)

		err = cleanDir(directory, "save_devel_*.bak")
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	log.Printf("end dump backup")
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
