package main

import (
	"encoding/base64"
	"net/http"

	"fmt"

	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/satori/go.uuid"
)

type User struct {
	Name         string
	Surname      string
	Username     string
	TempPassword bool
}

type SessionData struct {
	LoggedIn  bool
	SessionID string
	User      User
}

func clearSession(w http.ResponseWriter, r *http.Request) error {
	session, _ := cookieStore.Get(r, authCookieStoreName)
	sessionData := SessionData{}

	session.Values["SessionData"] = sessionData

	//save the session
	err := session.Save(r, w)
	if err != nil {
		return err
	}

	return nil
}

func createSession(w http.ResponseWriter, r *http.Request, user string, name string, surname string, tempPassword bool) (*SessionData, error) {
	session, _ := cookieStore.Get(r, authCookieStoreName)

	sessionID := uuid.NewV4()

	sessionData := SessionData{
		LoggedIn:  true,
		SessionID: sessionID.String(),
		User: User{
			Name:         name,
			Surname:      surname,
			Username:     user,
			TempPassword: tempPassword,
		},
	}

	err := saveSessionData(w, r, session, sessionData)
	if err != nil {
		return nil, err
	}

	return &sessionData, nil
}

func refreshSessionData(w http.ResponseWriter, r *http.Request, sessionData SessionData) error {
	session, _ := cookieStore.Get(r, authCookieStoreName)

	err := saveSessionData(w, r, session, sessionData)
	if err != nil {
		return err
	}

	return nil
}

func saveSessionData(w http.ResponseWriter, r *http.Request,
	session *sessions.Session, sessionData SessionData) error {

	session.Values["SessionData"] = sessionData

	//save the session
	err := session.Save(r, w)
	if err != nil {
		return err
	}

	return nil
}

func getSessionData(r *http.Request) (*SessionData, error) {
	session, _ := cookieStore.Get(r, authCookieStoreName)

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

	cookieStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   0,
		HttpOnly: true,
		Secure:   config.IsHTTPS,
	}

	return cookieStore, nil
}

func saveCookieEncodeKeys(keys [][]byte) error {
	query := dbUtils.PQuery(`
		INSERT INTO cookie_encode_key (
			encode_key,
			valid_from,
			valid_until
		)
		VALUES (?, ?, ?)
	`)

	dt := time.Now().UTC()
	after30days := dt.Add(time.Duration(30*24) * time.Hour)

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
		_, err = stmt.Exec(sKey, dt, after30days)
		if err != nil {
			return err
		}
	}

	tx.Commit()

	return nil
}

func getCookiesEncodeKeys() ([][]byte, error) {
	var keys [][]byte
	dt := time.Now().UTC()

	query := `
		SELECT encode_key
		  FROM cookie_encode_key
		 WHERE valid_from  <= ?
		   AND valid_until >= ?
		 ORDER BY cookie_encode_key_id
		 LIMIT 4
	`

	rows, err := db.Query(query, dt, dt)
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

	rows.Close()

	return keys, nil
}
