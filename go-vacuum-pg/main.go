package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
)

var (
	layout = "20060102"
)

func main() {
	var err error
	t := time.Now().UTC()
	sData := t.Format(layout)

	logFile, err := os.OpenFile(fmt.Sprintf("logs/vacuumlog_%s.txt", sData), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println(err)
		return
	}
	defer logFile.Close()

	mw := io.MultiWriter(os.Stdout, logFile)

	log.SetOutput(mw)

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

	var outb, errb bytes.Buffer

	cmd := exec.Command("vacuumdb", "-avzw", "-h", "devel", "-U", "postgres")
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(outb.String())
	log.Println(errb.String())

	log.Printf("*******************\nend vacuum\n")
}
