package main

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

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

		if len(user) == 0 || len(pass) == 0 {
			lres.bErr = true
			lres.sErr = "Unknown user or wrong password."
			loginError(user, "Unknown user or wrong password.")

			return &lres, nil
		}

		success, err := loginByUserPassword(user, pass)

		if err != nil || !success {
			lres.bErr = true
			lres.sErr = "Unknown user or wrong password."
			loginError(user, "Unknown user or wrong password.")

			return &lres, err
		}

		err = createSession(w, r, user)

		if err != nil {
			lres.bErr = true
			lres.sErr = err.Error()
			loginError(user, lres.sErr)

			return &lres, err
		}
	}

	lres.bErr = false
	loginSuccess(session.User.Username, "User logged in.")

	return &lres, nil
}

func loginSuccess(user string, msg string) {
	log.WithFields(logrus.Fields{
		"msg_type": "login",
		"status":   "successful",
		"user":     user,
	}).Info(msg)
}

func loginError(user string, msg string) {
	log.WithFields(logrus.Fields{
		"msg_type": "login",
		"status":   "failed",
		"user":     user,
	}).Error(msg)
}
