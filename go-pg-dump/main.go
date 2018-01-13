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
	t := time.Now().UTC()
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
		pg_restore -Fc -C save_devel_yyyymmdd.bak
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
