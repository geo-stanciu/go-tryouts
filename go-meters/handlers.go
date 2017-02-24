package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
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
	url := strings.ToLower(r.URL.Path)

	r.ParseForm()

	helper, err := getResponseHelperByURL(url)

	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"url": r.URL.Path,
		}).Error("Failed request")
	}

	if helper == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	model, err := helper.getResponse(w, r)

	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"url": r.URL.Path,
		}).Error("Failed request")
	}

	if model == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if model.Err() {
		setOperationError(w, r, model.SErr())
	}

	http.Redirect(w, r, model.Url(), http.StatusSeeOther)
}

func handleGetRequest(w http.ResponseWriter, r *http.Request) {
	url := strings.ToLower(r.URL.Path)

	session, err := getSessionData(r)

	if (err != nil || !session.LoggedIn) && url != "/login" {
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"url": r.URL.Path,
			}).Error("Failed request")
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

	bErr, sErr, err := getLastOperationError(w, r)

	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"url": r.URL.Path,
		}).Error("Failed request")
	}

	t := time.Now().Unix()

	passedObj := template0Data{
		Err:     bErr,
		SErr:    sErr,
		AppName: appName,
		Version: appVersion,
		Date:    t,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "private, max-age=600, no-store")

	page, err := getPageByURL(url)

	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"url": r.URL.Path,
		}).Error("Failed request")

		http.Error(w, fmt.Sprintf("%s - Not found", r.URL.Path), 404)
		return
	}

	if page == nil {
		http.Error(w, fmt.Sprintf("%s - Not found", r.URL.Path), 404)
		return
	}

	model, err := page.getModel(w, r)

	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"url": r.URL.Path,
		}).Error("Failed request")

		http.Error(w, fmt.Sprintf("%s - Not found", r.URL.Path), 404)
		return
	}

	if page.Template == "-" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	passedObj.Title = page.Title
	passedObj.Model = model
	passedObj.Session = *session

	err = executeTemplate(w, page.Template, passedObj)

	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"url": r.URL.Path,
		}).Error("Failed request")

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
