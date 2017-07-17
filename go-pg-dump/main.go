package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"time"
)

type Configuration struct {
	DumpDir string `json:"DumpDir"`
	DbName  string `json:"DbName"`
}

var (
	config = Configuration{}
)

func main() {
	var err error
	t := time.Now()
	sData := t.Format("20060102")

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

	dumpFile := path.Join(config.DumpDir, fmt.Sprintf("save_devel_%s.dmp", sData))

	log.Printf("start dump backup \"%s\"\n", dumpFile)

	dump, err := os.OpenFile(dumpFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Println(err)
		return
	}
	defer dump.Close()

	cmd := exec.Command("pg_dump", config.DbName)
	cmd.Stdout = dump

	err = cmd.Start()
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("end dump backup")
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
