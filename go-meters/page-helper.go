package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
)

type Page struct {
	Title      string
	Template   string
	Controller string
	Action     string
}

func (p *Page) getModel(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch p.Controller {
	case "Home":
		home := HomeController{}
		return p.getModelValue(home, w, r)

	default:
		return nil, nil
	}
}

func (p *Page) getModelValue(controller interface{}, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	if len(p.Action) == 0 || p.Action == "-" {
		return nil, nil
	}

	response := InvokeMethodByName(controller, p.Action, w, r, &ResponseHelper{})

	if len(response) >= 2 {
		i1 := response[0].Interface()
		i2 := response[1].Interface()

		if i2 != nil {
			return i1, i2.(error)
		}

		return i1, nil
	}

	return nil, fmt.Errorf("Function does not return the requested number of values.")
}

func getPageByURL(url string) (*Page, error) {
	var p Page
	var sURL string

	if url == "/" {
		sURL = "index"
	} else {
		sURL = strings.Replace(url[1:], ".html", "", 1)
	}

	query := `
        select page_title, 
               page_template,
               controller,
               action
          from wmeter.page
         where page_url = $1
    `

	err := db.QueryRow(query, sURL).Scan(&p.Title, &p.Template, &p.Controller, &p.Action)

	switch {
	case err == sql.ErrNoRows:
		err = fmt.Errorf("%s not found", url)
		return nil, err
	case err != nil:
		return nil, err
	}

	return &p, nil
}
