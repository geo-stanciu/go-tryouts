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
	HasURL() bool
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
		return res.getResponseValue(home, w, r)

	default:
		return nil, nil
	}
}

func (res *ResponseHelper) getResponseValue(controller interface{}, w http.ResponseWriter, r *http.Request) (ResponseModel, error) {
	if len(res.Action) == 0 || res.Action == "-" {
		return nil, nil
	}

	response := InvokeMethodByName(controller, res.Action, w, r, res)

	if len(response) >= 2 {
		r := response[0].Interface().(ResponseModel)
		i2 := response[1].Interface()

		if i2 != nil {
			return r, i2.(error)
		}

		return r, nil
	}

	return nil, fmt.Errorf("Function does not return the requested number of values.")
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
