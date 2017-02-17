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
	Title           string
	AppName         string
	Version         string
	Date            int64
	Model           interface{}
	HasResponseData bool
	ResponseModel   interface{}
}

type IndexModel struct {
	IsLoggedIn bool
	User       string
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		handleGetRequest(w, r)
	} else if r.Method == "POST" {
		handlePostRequest(w, r)
	} else {
		http.Error(w, "Method not allowed", 405)
		return
	}
}

func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	user := r.FormValue("username")
	//pass := r.FormValue("password")

	model := IndexModel{true, user}

	redirect2Url(w, r, "/index", model)
}

func handleGetRequest(w http.ResponseWriter, r *http.Request) {
	sendResponse(w, r, r.URL.Path, nil)
}

func redirect2Url(w http.ResponseWriter, r *http.Request, sURL string, responseModel interface{}) {
	sendResponse(w, r, sURL, responseModel)
}

func sendResponse(w http.ResponseWriter, r *http.Request, sURL string, responseModel interface{}) {
	url := strings.ToLower(sURL)

	if strings.HasSuffix(url, ".js") {
		http.ServeFile(w, r, sURL[1:])
		return
	}

	t := time.Now().Unix()

	passedObj := template0Data{
		AppName: appName,
		Version: appVersion,
		Date:    t,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "private, max-age=600, no-store")

	page, err := getPageByURL(url)

	if err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintf("%s - Not found", sURL), 404)
		return
	}

	if page == nil {
		http.Error(w, fmt.Sprintf("%s - Not found", sURL), 404)
		return
	}

	model, err := page.getModel(w, r)

	if err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintf("%s - Not found", sURL), 404)
		return
	}

	passedObj.Title = page.Title
	passedObj.Model = model

	if responseModel != nil {
		passedObj.HasResponseData = true
		passedObj.ResponseModel = responseModel
	} else {
		passedObj.HasResponseData = false
	}

	err = executeTemplate(w, page.Template, passedObj)

	if err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintf("%s - Not found", sURL), 404)
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
