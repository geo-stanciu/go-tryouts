package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
)

type Configuration struct {
	Port  string `json:"Port"`
	DbURL string `json:"DbURL"`
	Db    string `json:"DB"`
}

var (
	templateDelims   = []string{"{{%", "%}}"}
	templateBasePath string
	templates        *template.Template
	addr             *string
	db               *sql.DB
	config           = Configuration{}
)

func init() {
	// initialize the templates,
	// since we have custom delimiters.
	templateBasePath := "templates/"

	err := filepath.Walk(templateBasePath, parseTemplate)

	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	var err error

	cfgFile := "./conf.json"

	if _, err = os.Stat(cfgFile); os.IsNotExist(err) {
		log.Println(fmt.Sprintf("No config file was found with name: %s", cfgFile))
		os.Exit(1)
	}

	err = readConfig(cfgFile)

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
		http.StripPrefix("/images/", http.FileServer(http.Dir("resources/images"))))
	http.Handle("/js/",
		http.StripPrefix("/js/", http.FileServer(http.Dir("resources/js"))))
	http.Handle("/css/",
		http.StripPrefix("/css/", http.FileServer(http.Dir("resources/css"))))

	http.Handle("/favicon.ico", http.NotFoundHandler())

	http.HandleFunc("/", handler)

	err = http.ListenAndServe(*addr, nil)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func parseTemplate(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	// don't process folders themselves
	if info.IsDir() {
		return nil
	}

	templateName := path[len(templateBasePath):]

	if templates == nil {
		templates = template.New(templateName)
		templates.Delims(templateDelims[0], templateDelims[1])
		_, err = templates.ParseFiles(path)
	} else {
		_, err = templates.New(templateName).ParseFiles(path)
	}

	return err
}

func readConfig(cfgFile string) error {
	file, err := os.Open(cfgFile)

	if err != nil {
		return err
	}

	defer file.Close()

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&config)

	if err != nil {
		return err
	}

	return nil
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
