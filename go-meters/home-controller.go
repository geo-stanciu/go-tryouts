package main

import (
	"net/http"

	"./models"

	"strings"

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
			loginMessage(lres.BError, user, "Unknown user or wrong password.")

			return &lres, nil
		}

		success, err := ValidateUserPassword(user, pass)

		if err != nil || !success {
			lres.BError = true
			lres.SError = "Unknown user or wrong password."
			loginMessage(lres.BError, user, "Unknown user or wrong password.")

			return &lres, err
		}

		session, err = createSession(w, r, user)

		if err != nil {
			lres.BError = true
			lres.SError = err.Error()
			loginMessage(lres.BError, user, lres.SError)

			return &lres, err
		}
	}

	lres.BError = false
	loginMessage(lres.BError, session.User.Username, "User logged in.")

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

func (HomeController) Register(w http.ResponseWriter, r *http.Request) (*models.GenericResponseModel, error) {
	var lres models.GenericResponseModel

	user := r.FormValue("username")
	pass := r.FormValue("password")
	confirmPass := r.FormValue("confirm_password")
	name := r.FormValue("name")
	surname := r.FormValue("surname")
	email := r.FormValue("email")

	if len(user) == 0 {
		lres.BError = true
		lres.SError = "User is empty"
		registerMessage(lres.BError, user, email, lres.SError)

		return &lres, nil
	}

	if len(pass) == 0 || pass != confirmPass {
		lres.BError = true
		lres.SError = "Password is empty or is different from it's confirmation."
		registerMessage(lres.BError, user, email, lres.SError)

		return &lres, nil
	}

	if len(email) == 0 {
		lres.BError = true
		lres.SError = "E-mail is empty"
		registerMessage(lres.BError, user, email, lres.SError)

		return &lres, nil
	}

	u := MembershipUser{
		UserID:   -1,
		Username: user,
		Name:     name,
		Surname:  surname,
		Email:    email,
		Password: pass,
	}

	err := u.Save()

	if err != nil {
		lres.BError = true
		lres.SError = err.Error()
		registerMessage(lres.BError, user, email, lres.SError)

		return &lres, err
	}

	if isRequestFromLocalhost(r) && strings.ToLower(u.Username) == "admin" {
		err = u.AddToRole("Administrator")

		if err != nil {
			lres.BError = true
			lres.SError = err.Error()
			registerMessage(lres.BError, user, email, lres.SError)

			return &lres, err
		}
	}

	lres.BError = false
	lres.SError = "OK"
	registerMessage(lres.BError, user, email, lres.SError)

	return &lres, nil
}

func (HomeController) ChangePassword(w http.ResponseWriter, r *http.Request) (*models.GenericResponseModel, error) {
	var lres models.GenericResponseModel

	session, _ := getSessionData(r)

	if !session.LoggedIn {
		lres.BError = true
		lres.SError = "User not logged in."
		changePassMessage(lres.BError, "", "", lres.SError)

		return &lres, nil
	}

	var usr MembershipUser
	err := usr.GetUserByName(session.User.Username)

	if err != nil {
		lres.BError = true
		lres.SError = err.Error()
		changePassMessage(lres.BError, "", "", lres.SError)

		return &lres, nil
	}

	pass := r.FormValue("password")
	newPass := r.FormValue("new_password")
	confirmPass := r.FormValue("confirm_password")

	if len(pass) == 0 {
		lres.BError = true
		lres.SError = "Old password cannot be empty"
		changePassMessage(lres.BError, usr.Username, usr.Email, lres.SError)

		return &lres, nil
	}

	if len(newPass) == 0 || newPass != confirmPass {
		lres.BError = true
		lres.SError = "Password is empty or is different from it's confirmation."
		changePassMessage(lres.BError, usr.Username, usr.Email, lres.SError)

		return &lres, nil
	}

	if pass == newPass {
		lres.BError = true
		lres.SError = "The new password must be different from the current one."
		changePassMessage(lres.BError, usr.Username, usr.Email, lres.SError)

		return &lres, nil
	}

	valid, err := ValidateUserPassword(usr.Username, pass)

	if !valid {
		lres.BError = true
		lres.SError = "Old password is not valid."
		changePassMessage(lres.BError, usr.Username, usr.Email, lres.SError)

		return &lres, nil
	}

	usr.Password = newPass

	err = usr.Save()

	if err != nil {
		lres.BError = true
		lres.SError = err.Error()
		changePassMessage(lres.BError, usr.Username, usr.Email, lres.SError)

		return &lres, err
	}

	lres.BError = false
	lres.SError = "OK"
	changePassMessage(lres.BError, usr.Username, usr.Email, lres.SError)

	return &lres, nil
}

func loginMessage(isErr bool, user string, msg string) {
	if isErr {
		log.WithFields(logrus.Fields{
			"msg_type": "login",
			"status":   "failed",
			"user":     user,
		}).Error(msg)
	} else {
		log.WithFields(logrus.Fields{
			"msg_type": "login",
			"status":   "successful",
			"user":     user,
		}).Info(msg)
	}
}

func logoutMessage(user string, msg string) {
	log.WithFields(logrus.Fields{
		"msg_type": "logout",
		"user":     user,
	}).Info(msg)
}

func registerMessage(isErr bool, user string, email string, msg string) {
	if isErr {
		log.WithFields(logrus.Fields{
			"msg_type": "register",
			"status":   "failed",
			"user":     user,
			"email":    email,
		}).Error(msg)
	} else {
		log.WithFields(logrus.Fields{
			"msg_type": "register",
			"status":   "successful",
			"user":     user,
			"email":    email,
		}).Info(msg)
	}
}

func changePassMessage(isErr bool, user string, email string, msg string) {
	if isErr {
		log.WithFields(logrus.Fields{
			"msg_type": "change-password",
			"status":   "failed",
			"user":     user,
			"email":    email,
		}).Error(msg)
	} else {
		log.WithFields(logrus.Fields{
			"msg_type": "change-password",
			"status":   "successful",
			"user":     user,
			"email":    email,
		}).Info(msg)
	}
}
