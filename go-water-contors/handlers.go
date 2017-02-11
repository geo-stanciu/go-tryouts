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
	Title   string
	AppName string
	Version string
	Date    int64
	Model   interface{}
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	t := time.Now().Unix()

	passedObj := template0Data{
		Title:   "Index",
		AppName: appName,
		Version: appVersion,
		Date:    t,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "private, max-age=600, no-store")

	page, err := getPageByURL(strings.ToLower(r.URL.Path))

	if err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintf("%s - Not found", r.URL.Path), 404)
		return
	}

	if page == nil {
		http.Error(w, fmt.Sprintf("%s - Not found", r.URL.Path), 404)
		return
	}

	model, err := page.getModel()

	if err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintf("%s - Not found", r.URL.Path), 404)
		return
	}

	passedObj.Title = page.Title
	passedObj.Model = model

	err = executeTemplate(w, page.Template, passedObj)

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
