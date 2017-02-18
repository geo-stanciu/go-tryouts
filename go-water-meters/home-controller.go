package main

import "net/http"

type HomeController struct {
}

func (HomeController) Index() (interface{}, error) {
	return nil, nil
}

func (HomeController) Login(w http.ResponseWriter, r *http.Request) (*LoginResponse, error) {
	var lres LoginResponse

	session, _ := getSessionData(r)

	if !session.LoggedIn {
		user := r.FormValue("username")
		pass := r.FormValue("password")

		success, err := loginByUserPassword(user, pass)

		if err != nil || !success {
			lres.bErr = true
			lres.sErr = "Unknown user or wrong password."
			return &lres, err
		}

		err = createSession(w, r, user)

		if err != nil {
			lres.bErr = true
			lres.sErr = err.Error()
			return &lres, err
		}
	}

	lres.bErr = false

	return &lres, nil
}
