package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
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

	out, err := exec.Command("vacuumdb", "-avzwU", "postgres").Output()

	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(out))

	log.Printf("*******************\nend vacuum\n")
}
