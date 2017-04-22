package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"./models"
)

type ResponseHelper struct {
	Title           string
	Template        string
	Controller      string
	Action          string
	RedirectURL     string
	RedirectOnError string
}

func (res *ResponseHelper) getResponse(w http.ResponseWriter, r *http.Request) (models.ResponseModel, error) {
	switch res.Controller {
	case "Home":
		home := HomeController{}
		return res.getResponseValue(home, w, r)

	default:
		return nil, nil
	}
}

func (res *ResponseHelper) getResponseValue(controller interface{}, w http.ResponseWriter, r *http.Request) (models.ResponseModel, error) {
	if len(res.Action) == 0 || res.Action == "-" {
		return nil, nil
	}

	response := InvokeMethodByName(controller, res.Action, w, r, res)

	if len(response) >= 2 {
		r := response[0].Interface()
		i2 := response[1].Interface()

		if r == nil && i2 != nil {
			return nil, i2.(error)
		}

		if i2 != nil {
			return r.(models.ResponseModel), i2.(error)
		}

		return r.(models.ResponseModel), nil
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
        select request_title,
		       request_template,
		       controller,
               action,
			   redirect_url,
			   redirect_on_error
          from wmeter.request
         where request_url = $1
    `

	err := db.QueryRow(query, sURL).Scan(
		&res.Title,
		&res.Template,
		&res.Controller,
		&res.Action,
		&res.RedirectURL,
		&res.RedirectOnError,
	)

	switch {
	case err == sql.ErrNoRows:
		err = fmt.Errorf("%s not found", url)
		return nil, err
	case err != nil:
		return nil, err
	}

	return &res, nil
}
