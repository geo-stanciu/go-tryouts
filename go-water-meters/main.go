package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"encoding/gob"

	"github.com/gorilla/sessions"

	"strings"

	"github.com/gorilla/securecookie"
	_ "github.com/lib/pq"
)

var (
	templateDelims  = []string{"{{%", "%}}"}
	templates       *template.Template
	addr            *string
	db              *sql.DB
	config          = Configuration{}
	appName         = "Water Meter"
	appVersion      = "0.0.0.1"
	cookieStoreName = strings.Replace(appName, " ", "", -1)
	cookieStore     = sessions.NewCookieStore(
		securecookie.GenerateRandomKey(32),
		securecookie.GenerateRandomKey(32),
	)
)

func init() {
	gob.Register(&SessionData{})

	// initialize the templates,
	// since we have custom delimiters.
	basePath := "templates/"

	err := filepath.Walk(basePath, parseTemplate(basePath))

	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	var err error

	t := time.Now()
	sData := t.Format("20060102")

	logFile, err := os.OpenFile(fmt.Sprintf("logs/log_%s.txt", sData), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

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

	// server flags
	addr = flag.String("addr", ":"+config.Port, "http service address")

	flag.Parse()

	err = connect2Database(config.DbURL)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	defer db.Close()

	// Normal resources
	http.Handle("/static",
		http.FileServer(http.Dir("./static/")))
	http.Handle("/images/",
		http.StripPrefix("/images/", http.FileServer(http.Dir("public/images"))))
	http.Handle("/js/",
		http.StripPrefix("/js/", http.FileServer(http.Dir("public/js"))))
	http.Handle("/css/",
		http.StripPrefix("/css/", http.FileServer(http.Dir("public/css"))))

	http.Handle("/favicon.ico", http.NotFoundHandler())

	http.HandleFunc("/", handler)

	err = http.ListenAndServe(*addr, nil)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
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
