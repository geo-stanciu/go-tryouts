package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"encoding/gob"

	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"

	"strings"

	_ "github.com/lib/pq"
)

var (
	log             = logrus.New()
	templateDelims  = []string{"{{%", "%}}"}
	templates       *template.Template
	addr            *string
	db              *sql.DB
	config          = Configuration{}
	appName         = "GoMeters"
	appVersion      = "0.0.0.1"
	cookieStoreName = strings.Replace(appName, " ", "", -1)
	cookieStore     *sessions.CookieStore
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.Formatter = new(logrus.JSONFormatter)
	log.Level = logrus.DebugLevel

	// register SessionData for cookie use
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
	var auditLog AuditLog

	cfgFile := "./conf.json"
	err = config.ReadFromFile(cfgFile)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	err = connect2Database(config.DbURL)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	defer db.Close()

	mw := io.MultiWriter(os.Stdout, auditLog)
	log.Out = mw

	cookieStore, err = getNewCookieStore()

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// server flags
	addr = flag.String("addr", ":"+config.Port, "http service address")

	flag.Parse()

	log.WithField("port", *addr).Info("Starting listening...")

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

	log.Info("Closing application...")
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
