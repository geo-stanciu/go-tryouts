package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type template0Data struct {
	Version string
	Date    int64
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	t := time.Now().Unix()

	passedObj := template0Data{
		Version: "0.0.0.1",
		Date:    t,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "private, max-age=600, no-store")

	var filename string

	if r.URL.Path == "/" {
		filename = "index.html"
	} else if strings.Contains(r.URL.Path, ".html") {
		filename = fmt.Sprintf("%s", r.URL.Path[1:])
	} else {
		filename = fmt.Sprintf("%s.html", r.URL.Path[1:])
	}

	err := executeTemplate(w, filename, passedObj)

	if err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintf("%s - Not found", r.URL.Path), 404)
		return
	}
}

func executeTemplate(w io.Writer, tmplName string, data interface{}) error {
	var err error

	t := templates.Lookup(tmplName)

	if t == nil {
		errNoLayout := fmt.Errorf("%s not found", tmplName)
		return errNoLayout
	}

	layout := templates.Lookup("layout")

	if layout == nil {
		errNoLayout := errors.New("layout.html not found")
		return errNoLayout
	}

	page, err := layout.Clone()

	if err != nil {
		return err
	}

	_, err = page.AddParseTree("content", t.Tree)

	if err != nil {
		return err
	}

	return page.Execute(w, data)
}
