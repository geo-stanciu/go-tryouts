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
	sync.RWMutex
	UserID      int
	Username    string
	Name        string
	Surname     string
	Email       string
	Password    string
	NewUsername string
}

func (u *MembershipUser) UserExists(user string) (bool, error) {
	u.RLock()
	defer u.RUnlock()

	found := false

	query := `
        SELECT EXISTS(
			SELECT *
		      FROM wmeter.user
	         WHERE loweredusername = lower($1)
		)
    `

	err := db.QueryRow(query, user).Scan(&found)

	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, err
	}

	return found, nil
}

func (u *MembershipUser) GetUserByName(user string) error {
	u.Lock()
	defer u.Unlock()

	query := `
        SELECT user_id
		       username,
               name,
               surname,
               email
          FROM wmeter.user
         WHERE loweredusername = lower($1)
    `

	err := db.QueryRow(query, user).Scan(
		&u.UserID,
		&u.Username,
		&u.Name,
		&u.Surname,
		&u.Email)

	switch {
	case err == sql.ErrNoRows:
		return fmt.Errorf("username \"%s\" not found", user)
	case err != nil:
		return err
	}

	return nil
}

func LoginByUserPassword(user string, pass string) (bool, error) {
	var hashedPassword string
	var passwordSalt string

	query := `
        SELECT COALESCE(p.password, '') AS password,
		       COALESCE(p.password_salt, '') AS password_salt
          FROM wmeter.user u
          LEFT OUTER JOIN wmeter.user_password p ON (u.user_id = p.user_id)
         WHERE loweredusername = lower($1)
		   AND p.valid_from >= current_timestamp
		   AND (p.valid_until is null OR p.valid_until > current_timestamp)
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

func (u *MembershipUser) testSaveUser(tx *sql.Tx) error {
	if len(u.Username) == 0 {
		return fmt.Errorf("unknown user \"%s\"", u.Username)
	}

	if u.UserID <= 0 && len(u.Password) == 0 {
		return fmt.Errorf("cannot create user with empty password")
	}

	var found bool

	query := `
        SELECT EXISTS(
			SELECT *
		      FROM wmeter.user
			 WHERE loweredusername = LOWER($1)
			   AND user_id <> $2
		)
	`

	stmt, err := tx.Prepare(query)

	if err != nil {
		return err
	}

	defer stmt.Close()

	err = stmt.QueryRow(u.Username, u.UserID).Scan(&found)

	switch {
	case err == sql.ErrNoRows:
		found = false
	case err != nil:
		return err
	}

	if found {
		return fmt.Errorf("duplicate user \"%s\"", u.Username)
	}

	if len(u.NewUsername) > 0 {
		err = stmt.QueryRow(u.NewUsername, u.UserID).Scan(&found)

		switch {
		case err == sql.ErrNoRows:
			found = false
		case err != nil:
			return err
		}

		if found {
			return fmt.Errorf("duplicate user \"%s\"", u.NewUsername)
		}
	}

	return nil
}

func (u *MembershipUser) Save() error {
	var query string

	u.Lock()
	defer u.Unlock()

	tx, err := db.Begin()

	if err != nil {
		return err
	}

	defer tx.Rollback()

	err = u.testSaveUser(tx)

	if err != nil {
		return err
	}

	if u.UserID <= 0 {
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
		`

		_, err = tx.Exec(
			query,
			u.Username,
			strings.ToLower(u.Username),
			u.Name,
			u.Surname,
			u.Email,
			strings.ToLower(u.Email),
		)

		query = `
			SELECT user_id FROM wmeter.user WHERE loweredusername = $1
		`

		err = tx.QueryRow(query, strings.ToLower(u.Username)).Scan(&u.UserID)

		switch {
		case err == sql.ErrNoRows:
			u.UserID = -1
		case err != nil:
			return err
		}

		if u.UserID <= 0 {
			return fmt.Errorf("unknown user \"%s\"", u.Username)
		}

		err = u.ChangePassword(tx)

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
				   loweredemail    = $6,
				   last_update     = current_timestamp
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
			u.UserID,
		)

		if err != nil {
			return err
		}

		if len(u.Password) > 0 {
			err = u.ChangePassword(tx)

			if err != nil {
				return err
			}
		}
	}

	tx.Commit()

	return nil
}

func (u *MembershipUser) ChangePassword(tx *sql.Tx) error {
	var passwordID int
	var oldPassword string
	var oldSalt string
	var validityDate NullTime

	saltBytes := uuid.NewV4()
	salt := saltBytes.String()

	passwordBytes := []byte(salt + u.Password)
	password := string(passwordBytes)

	query := `
		SELECT password_id,
		       password,
			   password_salt,
			   valid_until
		  FROM wmeter.user_password
		 WHERE user_id = $1
	`

	err := tx.QueryRow(query, u.UserID).Scan(
		&passwordID,
		&oldPassword,
		&oldSalt,
		&validityDate)

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

	_, err = tx.Exec(query, u.UserID, password, salt)

	if err != nil {
		return err
	}

	return nil
}
