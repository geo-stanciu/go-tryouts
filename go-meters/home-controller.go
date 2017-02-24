package main

import (
	"net/http"

	"./models"

	"github.com/sirupsen/logrus"
)

type HomeController struct {
}

func (HomeController) Index() (interface{}, error) {
	return nil, nil
}

func (HomeController) Login(w http.ResponseWriter, r *http.Request) (*models.LoginResponse, error) {
	var lres models.LoginResponse

	session, _ := getSessionData(r)

	if !session.LoggedIn {
		user := r.FormValue("username")
		pass := r.FormValue("password")

		if len(user) == 0 || len(pass) == 0 {
			lres.BError = true
			lres.SError = "Unknown user or wrong password."
			loginError(user, "Unknown user or wrong password.")

			return &lres, nil
		}

		success, err := LoginByUserPassword(user, pass)

		if err != nil || !success {
			lres.BError = true
			lres.SError = "Unknown user or wrong password."
			loginError(user, "Unknown user or wrong password.")

			return &lres, err
		}

		session, err = createSession(w, r, user)

		if err != nil {
			lres.BError = true
			lres.SError = err.Error()
			loginError(user, lres.SError)

			return &lres, err
		}
	}

	lres.BError = false
	loginSuccess(session.User.Username, "User logged in.")

	return &lres, nil
}

func (HomeController) Logout(w http.ResponseWriter, r *http.Request) (*models.GenericResponseModel, error) {
	var lres models.GenericResponseModel
	var user string

	session, _ := getSessionData(r)

	if session.LoggedIn {
		user = session.User.Username
		err := clearSession(w, r)

		if err != nil {
			lres.BError = true
			lres.SError = err.Error()
			return nil, err
		}
	}

	logoutMessage(user, "User logged out.")

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

func logoutMessage(user string, msg string) {
	log.WithFields(logrus.Fields{
		"msg_type": "logout",
		"user":     user,
	}).Info(msg)
}
