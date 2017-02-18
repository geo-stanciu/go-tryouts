package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
)

type ResponseModel interface {
	Err() bool
	SErr() string
	Url() string
	SetURL(string)
}

type ResponseHelper struct {
	Controller      string
	Action          string
	RedirectURL     string
	RedirectOnError string
}

func (res *ResponseHelper) getResponse(w http.ResponseWriter, r *http.Request) (ResponseModel, error) {
	switch res.Controller {
	case "Home":
		home := HomeController{}

		switch res.Action {
		case "Login":
			r, err := home.Login(w, r)

			if r.Err() {
				r.SetURL(res.RedirectOnError)
			} else {
				r.SetURL(res.RedirectURL)
			}

			return r, err

		default:
			return nil, nil
		}

	default:
		return nil, nil
	}
}

func getResponseHelperByURL(url string) (*ResponseHelper, error) {
	var res ResponseHelper
	var sURL string

	if url == "/" {
		sURL = "index"
	} else {
		sURL = strings.Replace(url[1:], ".html", "", 1)
	}

	query := `
        select controller,
               action,
			   redirect_url,
			   redirect_on_error
          from wmeter.request
         where request_url = $1
    `

	err := db.QueryRow(query, sURL).Scan(&res.Controller, &res.Action, &res.RedirectURL, &res.RedirectOnError)

	switch {
	case err == sql.ErrNoRows:
		err = fmt.Errorf("%s not found", url)
		return nil, err
	case err != nil:
		return nil, err
	}

	return &res, nil
}
