package main

import (
	"database/sql"
	"fmt"
	"sync"

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
	var hashedPassword string
	var passwordSalt string

	query := `
        SELECT p.password, p.password_salt
          FROM wmeter.user u
          LEFT OUTER JOIN wmeter.user_password p ON (u.user_id = p.user_id)
         WHERE lower(u.username) = lower($1)
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

func (u *MembershipUser) Save() error {
	var userID int
	var passwordID int
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
		 WHERE lower(username) = lower($1)
	`

	err = tx.QueryRow(query, u.Username).Scan(&userID)

	switch {
	case err == sql.ErrNoRows:
		userID = -1
	case err != nil:
		return err
	}

	if userID <= 0 {
		query := `
			INSERT INTO wmeter.user (
				username,
				name,
				surname,
				email
			)
			VALUES (
				$1, $2, $3, $4
			)
			RETURNING user_id
		`

		err = tx.QueryRow(query, u.Username, u.Name, u.Surname, u.Email).Scan(&userID)

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
			   SET username = $1,
			       name     = $2,
				   surname  = $3,
				   email    = $4
			 WHERE user_id = $5
		`

		usr := u.Username

		if len(u.NewUsername) > 0 {
			usr = u.NewUsername
		}

		_, err = tx.Exec(query, usr, u.Name, u.Surname, u.Email, userID)

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
				AND valid_until is not null
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
