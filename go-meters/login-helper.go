package main

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/satori/go.uuid"
)

type User struct {
	Name     string
	Surname  string
	Username string
}

type SessionData struct {
	LoggedIn  bool
	SessionID string
	User      User
}

type LoginResponse struct {
	bErr bool
	sErr string
	URL  string
}

func (l *LoginResponse) Err() bool {
	return l.bErr
}

func (l *LoginResponse) SErr() string {
	return l.sErr
}

func (l *LoginResponse) Url() string {
	return l.URL
}

func (l *LoginResponse) SetURL(url string) {
	l.URL = url
}

func createSession(w http.ResponseWriter, r *http.Request, user string) error {
	session, _ := cookieStore.Get(r, cookieStoreName)

	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   0,
		HttpOnly: true,
		//Secure: true // for https
	}

	sessionID := uuid.NewV4()

	sessionData := SessionData{
		LoggedIn:  true,
		SessionID: sessionID.String(),
		User: User{
			Name:     "name1",
			Surname:  "surname1",
			Username: user,
		},
	}

	session.Values["SessionData"] = sessionData

	//save the session
	err := session.Save(r, w)

	if err != nil {
		return err
	}

	return nil
}

func getSessionData(r *http.Request) (*SessionData, error) {
	session, _ := cookieStore.Get(r, cookieStoreName)

	// Retrieve our struct and type-assert it
	val := session.Values["SessionData"]
	var sessionData = &SessionData{}

	if val == nil {
		return sessionData, nil
	}

	data, ok := val.(*SessionData)

	if !ok {
		return nil, nil
	}

	return data, nil
}
