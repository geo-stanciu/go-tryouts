package main

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type MembershipUser struct {
	UserID   int
	Username string
	Name     string
	Surname  string
	Email    string
}

func getUserByName(user string) (*MembershipUser, error) {
	var u MembershipUser

	query := `
        SELECT user_id,
               username,
               name,
               surname,
               email
          FROM wmeter.user
         WHERE lower(username) = lower($1)
    `

	err := db.QueryRow(query, user).Scan(&u.UserID, &u.Username, &u.Name, &u.Surname, &u.Email)

	switch {
	case err == sql.ErrNoRows:
		return nil, fmt.Errorf("username \"%s\" not found", user)
	case err != nil:
		return nil, err
	}

	return &u, nil
}

func loginByUserPassword(user string, pass string) (bool, error) {
	var password string
	var passwordSalt string

	query := `
        SELECT p.password, p.password_salt
          FROM wmeter.user u
          LEFT OUTER JOIN wmeter.user_password p ON u.user_id = p.user_id
         WHERE lower(u.username) = lower($1)
    `

	err := db.QueryRow(query, user).Scan(&password, &passwordSalt)

	switch {
	case err == sql.ErrNoRows:
		return false, fmt.Errorf("username \"%s\" not found", user)
	case err != nil:
		return false, err
	}

	passBytes := []byte(passwordSalt + pass)

	hashedPassword, err := bcrypt.GenerateFromPassword(passBytes, bcrypt.DefaultCost)

	if err != nil {
		return false, err
	}

	computedPassword := string(hashedPassword)

	if password == computedPassword {
		return true, nil
	}

	return false, nil
}
