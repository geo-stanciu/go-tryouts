package main

import (
	"net/http"

	"github.com/gorilla/sessions"
)

func setOperationError(w http.ResponseWriter, r *http.Request, sError string) error {
	session, _ := cookieStore.Get(r, cookieStoreName)

	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   0,
		HttpOnly: true,
		//Secure: true // for https
	}

	session.Values["Err"] = true
	session.Values["SErr"] = sError

	//save the session
	err := session.Save(r, w)

	if err != nil {
		return err
	}

	return nil
}

func getLastOperationError(w http.ResponseWriter, r *http.Request) (bool, string, error) {
	session, _ := cookieStore.Get(r, cookieStoreName)

	vErr := session.Values["Err"]
	bErr, ok := vErr.(bool)

	if !ok {
		return false, "", nil
	}

	vSErr := session.Values["SErr"]
	sErr, ok := vSErr.(string)

	if !ok {
		return false, "", nil
	}

	// clear last err
	session.Values["Err"] = false
	session.Values["SErr"] = ""

	//save the session
	err := session.Save(r, w)

	if err != nil {
		return false, "", err
	}

	return bErr, sErr, nil
}
