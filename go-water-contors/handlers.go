package main

import (
	"fmt"
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

	var filename string

	if r.URL.Path == "/" {
		filename = "index.html"
	} else if strings.Contains(r.URL.Path, ".html") {
		filename = fmt.Sprintf("%s", r.URL.Path[1:])
	} else {
		filename = fmt.Sprintf("%s.html", r.URL.Path[1:])
	}

	err := templates.ExecuteTemplate(w, filename, passedObj)

	if err != nil {
		log.Println(err)
		http.Error(w, "Not found", 404)
		return
	}
}
