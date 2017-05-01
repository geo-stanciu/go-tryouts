package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
)

type Configuration struct {
	DbName string `json:"DbName"`
}

var (
	config = Configuration{}
)

func main() {
	var err error
	t := time.Now()
	sData := t.Format("20060102")

	logFile, err := os.OpenFile(fmt.Sprintf("logs/vacuumlog_%s.txt", sData), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
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

	out, err := exec.Command(
		"psql",
		"-U",
		"postgres",
		"-d",
		config.DbName,
		"-c",
		"vacuum analyse verbose;",
	).Output()

	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(out))

	log.Printf("*******************\nend vacuum\n")
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
