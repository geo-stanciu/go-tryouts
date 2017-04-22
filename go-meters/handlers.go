package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type template0Data struct {
	Err     bool
	SErr    string
	Title   string
	AppName string
	Version string
	Date    int64
	Session SessionData
	Model   interface{}
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
	url := getBaseURL(r)
	session, err := getSessionData(r)

	if (err != nil || !session.LoggedIn) && url != "/perform-login" && url != "/perform-register" {
		if err != nil {
			Log(true, err, "no-context", "Failed request", "url", r.URL.Path)
		}

		setOperationError(w, r, "Request failed.")

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if session.LoggedIn && strings.HasPrefix(url, "/perform-login") {
		setOperationError(w, r, "Request failed.")

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	r.ParseForm()

	handleRequest(w, r, url, session)
}

func handleGetRequest(w http.ResponseWriter, r *http.Request) {
	url := getBaseURL(r)
	session, err := getSessionData(r)

	if (err != nil || !session.LoggedIn) && url != "/login" && url != "/register" {
		if err != nil {
			Log(true, err, "no-context", "Failed request", "url", r.URL.Path)
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if session.LoggedIn && strings.HasPrefix(url, "/login") {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if strings.HasSuffix(url, ".js") {
		http.ServeFile(w, r, r.URL.Path[1:])
		return
	}

	handleRequest(w, r, url, session)
}

func handleRequest(w http.ResponseWriter, r *http.Request, url string, session *SessionData) {
	bErr, sErr, err := getLastOperationError(w, r)

	if err != nil {
		Log(true, err, "no-context", "Failed request", "url", r.URL.Path)
	}

	t := time.Now().Unix()

	passedObj := template0Data{
		Err:     bErr,
		SErr:    sErr,
		AppName: appName,
		Version: appVersion,
		Date:    t,
	}

	response, err := getResponseHelperByURL(url)

	if err != nil {
		Log(true, err, "no-context", "Failed request", "url", r.URL.Path)

		http.Error(w, fmt.Sprintf("%s - Not found", r.URL.Path), 404)
		return
	}

	if response == nil {
		http.Error(w, fmt.Sprintf("%s - Not found", r.URL.Path), 404)
		return
	}

	model, err := response.getResponse(w, r)

	if err != nil {
		Log(true, err, "no-context", "Failed request", "url", r.URL.Path)
	}

	passedObj.Title = response.Title
	passedObj.Model = model
	passedObj.Session = *session

	if response.Template != "-" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "private, max-age=600, no-store")

		err = executeTemplate(w, response.Template, passedObj)

		if err != nil {
			Log(true, err, "no-context", "Failed request", "url", r.URL.Path)

			http.Error(w, fmt.Sprintf("%s - Not found", r.URL.Path), 404)
			return
		}

		return
	}

	if model.Err() {
		setOperationError(w, r, model.SErr())
	} else {
		setOperationSuccess(w, r, model.SErr())
	}

	if model.HasURL() {
		http.Redirect(w, r, model.Url(), http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(model)
	if err != nil {
		setOperationError(w, r, err.Error())
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

func getBaseURL(r *http.Request) string {
	url := strings.ToLower(r.URL.Path)
	idx := getEndIdxOfBaseURL(url)

	if len(url) > 0 && idx > 0 {
		url = url[0:idx]
	}

	return url
}

func getEndIdxOfBaseURL(url string) int {
	lastSlash := strings.LastIndex(url, "/")
	lastQ := strings.LastIndex(url, "?")
	lastHash := strings.LastIndex(url, "#")

	idx := getMinGreaterThanZero(lastSlash, lastQ)
	idx = getMinGreaterThanZero(idx, lastHash)

	return idx
}

func getMinGreaterThanZero(a, b int) int {
	if a > 0 && b > 0 {
		if a <= b {
			return a
		}
		return b
	} else if a > 0 {
		return a
	} else if b > 0 {
		return b
	}

	return -1
}
