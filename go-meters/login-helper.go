package main

import (
	"encoding/base64"
	"net/http"

	"fmt"

	"github.com/gorilla/securecookie"
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

func createSession(w http.ResponseWriter, r *http.Request, user string) (*SessionData, error) {
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
		return nil, err
	}

	return &sessionData, nil
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

func getNewCookieStore() (*sessions.CookieStore, error) {
	encodeKeys, err := getCookiesEncodeKeys()

	if err != nil {
		return nil, err
	}

	if encodeKeys == nil {
		key1 := securecookie.GenerateRandomKey(32)
		key2 := securecookie.GenerateRandomKey(32)
		encodeKeys = append(encodeKeys, key1)
		encodeKeys = append(encodeKeys, key2)

		err = saveCookieEncodeKeys(encodeKeys)

		if err != nil {
			return nil, err
		}
	}

	length := len(encodeKeys)

	if length == 0 {
		return nil, fmt.Errorf("Could not find any cookie encode keys")
	}

	var cookieStore *sessions.CookieStore

	if length >= 4 {
		cookieStore = sessions.NewCookieStore(
			encodeKeys[0],
			encodeKeys[1],
			encodeKeys[2],
			encodeKeys[3],
		)
	} else if len(encodeKeys) >= 2 {
		cookieStore = sessions.NewCookieStore(
			encodeKeys[0],
			encodeKeys[1],
		)
	} else {
		cookieStore = sessions.NewCookieStore(encodeKeys[0])
	}

	return cookieStore, nil
}

func saveCookieEncodeKeys(keys [][]byte) error {
	query := `
		INSERT INTO wmeter.cookie_encode_key (
			encode_key
		)
		VALUES (
			$1
		)
	`

	tx, err := db.Begin()

	if err != nil {
		return err
	}

	defer tx.Rollback()

	stmt, err := tx.Prepare(query)

	if err != nil {
		return err
	}

	defer stmt.Close()

	for _, key := range keys {
		sKey := base64.StdEncoding.EncodeToString(key)
		_, err = stmt.Exec(sKey)
	}

	tx.Commit()

	return nil
}

func getCookiesEncodeKeys() ([][]byte, error) {
	var keys [][]byte

	query := `
		SELECT encode_key
		  FROM wmeter.cookie_encode_key
		 WHERE valid_from  <= current_timestamp
		   AND valid_until >= current_timestamp
		 ORDER BY cookie_encode_key_id
		 LIMIT 4
	`

	rows, err := db.Query(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var encodeKey string
		err = rows.Scan(&encodeKey)

		if err != nil {
			return nil, err
		}

		key, err := base64.StdEncoding.DecodeString(encodeKey)

		if err != nil {
			return nil, err
		}

		keys = append(keys, key)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return keys, nil
}
