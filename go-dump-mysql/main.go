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

	"github.com/geo-stanciu/go-utils/utils"
)

type configuration struct {
	DumpDir    string   `json:"DumpDir"`
	Files2Keep int      `json:"Files2Keep"`
	User       string   `json:"User"`
	Password   string   `json:"Password"`
	DbNames    []string `json:"DbNames"`
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

	for _, dbname := range config.DbNames {
		dumpname := fmt.Sprintf("save_%s_%s.sql", dbname, sData)
		dumpFile := path.Join(config.DumpDir, dumpname)
		zipname := path.Join(config.DumpDir, fmt.Sprintf("save_%s_%s.zip", dbname, sData))

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
			dbname,
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

		directory := getAbsPath(config.DumpDir)

		if config.Files2Keep > 0 {
			log.Printf("\n\nCleaning old files from \"%s\"\n", directory)
			log.Printf("Will keep the last %d files.", config.Files2Keep)

			err = cleanDir(directory, fmt.Sprintf("save_%s_*.zip", dbname))
			if err != nil {
				log.Fatal(err)
				return
			}
		}
	}

	log.Printf("\n\nend dump backup")
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
