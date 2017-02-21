package main

import (
	"database/sql"
	"fmt"
	"sync"

	"strings"

	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type MembershipUser struct {
	Username    string
	Name        string
	Surname     string
	Email       string
	Password    string
	NewUsername string
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
         WHERE loweredusername = lower($1)
    `

	err := db.QueryRow(query, user).Scan(
		&u.Username,
		&u.Name,
		&u.Surname,
		&u.Email)

	switch {
	case err == sql.ErrNoRows:
		return nil, fmt.Errorf("username \"%s\" not found", user)
	case err != nil:
		return nil, err
	}

	return &u, nil
}

func loginByUserPassword(user string, pass string) (bool, error) {
	var hashedPassword string
	var passwordSalt string

	query := `
        SELECT p.password, p.password_salt
          FROM wmeter.user u
          LEFT OUTER JOIN wmeter.user_password p ON (u.user_id = p.user_id)
         WHERE loweredusername = lower($1)
    `

	err := db.QueryRow(query, user).Scan(&hashedPassword, &passwordSalt)

	switch {
	case err == sql.ErrNoRows:
		return false, fmt.Errorf("username \"%s\" not found", user)
	case err != nil:
		return false, err
	}

	passBytes := []byte(passwordSalt + pass)
	hashBytes := []byte(hashedPassword)

	err = bcrypt.CompareHashAndPassword(hashBytes, passBytes)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (u *MembershipUser) testSaveUser(tx *sql.Tx, userID int) error {
	if len(u.Username) == 0 {
		return fmt.Errorf("unknown user \"%s\"", u.Username)
	}

	if userID <= 0 && len(u.Password) == 0 {
		return fmt.Errorf("cannot create user with empty password")
	}

	var contor int

	query := `
        SELECT COUNT(*)
         FROM wmeter.user
        WHERE loweredusername = LOWER($1)
        AND user_id <> $2
	`

	stmt, err := tx.Prepare(query)

	if err != nil {
		return err
	}

	defer stmt.Close()

	err = stmt.QueryRow(u.Username, userID).Scan(&contor)

	switch {
	case err == sql.ErrNoRows:
		contor = 0
	case err != nil:
		return err
	}

	if contor > 0 {
		return fmt.Errorf("duplicate user \"%s\"", u.Username)
	}

	if len(u.NewUsername) > 0 {
		err = stmt.QueryRow(u.NewUsername, userID).Scan(&contor)

		switch {
		case err == sql.ErrNoRows:
			contor = 0
		case err != nil:
			return err
		}

		if contor > 0 {
			return fmt.Errorf("duplicate user \"%s\"", u.NewUsername)
		}
	}

	return nil
}

func (u *MembershipUser) Save() error {
	var userID int
	var mutex = &sync.Mutex{}

	mutex.Lock()
	defer mutex.Unlock()

	tx, err := db.Begin()

	if err != nil {
		return err
	}

	defer tx.Rollback()

	query := `
		SELECT user_id
		  FROM wmeter.user
		 WHERE loweredusername = lower($1)
	`

	err = tx.QueryRow(query, u.Username).Scan(&userID)

	switch {
	case err == sql.ErrNoRows:
		userID = -1
	case err != nil:
		return err
	}

	err = u.testSaveUser(tx, userID)

	if err != nil {
		return err
	}

	if userID <= 0 {
		query := `
			INSERT INTO wmeter.user (
				username,
				loweredusername,
				name,
				surname,
				email,
				loweredemail
			)
			VALUES (
				$1, $2, $3, $4, $5, $6
			)
			RETURNING user_id
		`

		err = tx.QueryRow(
			query,
			u.Username,
			strings.ToLower(u.Username),
			u.Name,
			u.Surname,
			u.Email,
			strings.ToLower(u.Email),
		).Scan(&userID)

		switch {
		case err == sql.ErrNoRows:
			userID = -1
		case err != nil:
			return err
		}

		if userID <= 0 {
			return fmt.Errorf("unknown user \"%s\"", u.Username)
		}

		err = savePassword(tx, userID, u.Password)

		if err != nil {
			return err
		}
	} else {
		query = `
			UPDATE wmeter.user
			   SET username        = $1,
			       loweredusername = $2,
			       name            = $3,
				   surname         = $4,
				   email           = $5,
				   loweredemail    = $6
			 WHERE user_id = $7
		`

		usr := u.Username

		if len(u.NewUsername) > 0 {
			usr = u.NewUsername
		}

		_, err = tx.Exec(
			query,
			usr,
			strings.ToLower(usr),
			u.Name,
			u.Surname,
			u.Email,
			strings.ToLower(u.Email),
			userID,
		)

		if err != nil {
			return err
		}

		if len(u.Password) > 0 {
			err = savePassword(tx, userID, u.Password)

			if err != nil {
				return err
			}
		}
	}

	tx.Commit()

	return nil
}

func savePassword(tx *sql.Tx, userID int, pass string) error {
	var passwordID int
	var oldPassword string
	var oldSalt string
	var validityDate NullTime

	saltBytes := uuid.NewV4()
	salt := saltBytes.String()

	passwordBytes := []byte(salt + pass)
	password := string(passwordBytes)

	query := `
		SELECT password_id,
		       password,
			   password_salt,
			   valid_until
		  FROM wmeter.user_password
		 WHERE user_id = $1
	`

	err := tx.QueryRow(query, userID).Scan(&passwordID, &oldPassword, &oldSalt, &validityDate)

	switch {
	case err == sql.ErrNoRows:
		passwordID = -1
	case err != nil:
		return err
	}

	if passwordID > 0 {
		if oldPassword == password && oldSalt == salt {
			return nil
		}

		query = `
			UPDATE wmeter.user_password
			   SET valid_until = statement_timestamp()
			 WHERE password_id = $1
			   AND valid_until is null
		`

		_, err = tx.Exec(query, passwordID)

		if err != nil {
			return err
		}

		query = `
			INSERT INTO wmeter.user_password_archive (
				password_id,
				user_id,
				password,
				password_salt,
				valid_from,
				valid_until
			)
			SELECT password_id,
					user_id,
					password,
					password_salt,
					valid_from,
					valid_until
				FROM wmeter.user_password
				WHERE password_id = $1
		`

		_, err = tx.Exec(query, passwordID)

		if err != nil {
			return err
		}

		query = `
			DELETE FROM wmeter.user_password WHERE password_id = $1
		`

		_, err = tx.Exec(query, passwordID)

		if err != nil {
			return err
		}
	}

	query = `
		INSERT INTO wmeter.user_password (
			user_id,
			password,
			password_salt
		)
		VALUES (
			$1, $2, $3
		)
	`

	_, err = tx.Exec(query, userID, password, salt)

	if err != nil {
		return err
	}

	return nil
}
