package main

import "database/sql"
import "strings"

import "fmt"

type Page struct {
	Title      string
	Template   string
	Controller string
	Action     string
}

func (p Page) getModel() (interface{}, error) {
	switch p.Controller {
	case "Home":
		home := HomeController{}

		switch p.Action {
		case "Index":
			return home.Index()

		default:
			return nil, nil
		}

	default:
		return nil, nil
	}
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
