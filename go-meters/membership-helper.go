package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"sync"

	"strings"

	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type MembershipUser struct {
	sync.RWMutex
	UserID   int
	Username string
	Name     string
	Surname  string
	Email    string
	Password string
}

func (u *MembershipUser) UserExists(user string) (bool, error) {
	u.RLock()
	defer u.RUnlock()

	found := false

	query := `
        SELECT EXISTS(
			SELECT 1
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
        SELECT user_id,
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

func (u *MembershipUser) GetUserByID(userID int) error {
	u.Lock()
	defer u.Unlock()

	query := `
        SELECT user_id,
		       username,
               name,
               surname,
               email
          FROM wmeter.user
         WHERE user_id = $1
    `

	err := db.QueryRow(query, userID).Scan(
		&u.UserID,
		&u.Username,
		&u.Name,
		&u.Surname,
		&u.Email)

	switch {
	case err == sql.ErrNoRows:
		return fmt.Errorf("username not found")
	case err != nil:
		return err
	}

	return nil
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
			SELECT 1
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

		if err != nil {
			return err
		}

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

		log.WithFields(logrus.Fields{
			"msg_type": "add-user",
			"status":   "successful",
			"new":      u,
		}).Info("Add new user.")
	} else {
		var old MembershipUser
		err = old.GetUserByID(u.UserID)

		if err != nil {
			return err
		}

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

		_, err = tx.Exec(
			query,
			u.Username,
			strings.ToLower(u.Username),
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

		log.WithFields(logrus.Fields{
			"msg_type": "update-user",
			"status":   "successful",
			"old":      old,
			"new":      u,
		}).Info("Update user.")
	}

	tx.Commit()

	return nil
}

func (u *MembershipUser) GetUserRoles() ([]MembershipRole, error) {
	u.RLock()
	defer u.RUnlock()

	var roles []MembershipRole

	query := `
		SELECT r.role_id,
		       r.role
	      FROM wmeter.user_role ur
		  JOIN wmeter.role r ON (ur.role_id = r.role_id)
		 WHERE ur.user_id =  $1
		   AND ur.role_id =  $2
		   AND ur.valid_from     <= current_timestamp
		   AND (ur.valid_until is null OR ur.valid_until > current_timestamp)
		 ORDER BY r.role
	`

	rows, err := db.Query(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var r MembershipRole
		err = rows.Scan(&r.RoleID, &r.Rolename)

		if err != nil {
			return nil, err
		}

		roles = append(roles, r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

func (u *MembershipUser) AddToRole(role string) error {
	u.Lock()
	defer u.Unlock()

	var r MembershipRole
	err := r.GetRoleByName(role)

	if err != nil {
		return err
	}

	found, err := r.HasMemberID(u.UserID)

	if err != nil {
		return err
	}

	if found {
		return nil
	}

	query := `
		INSERT INTO wmeter.user_role (
			user_id,
			role_id
		)
		VALUES (
			$1, $2
		)
	`

	_, err = db.Exec(
		query,
		u.UserID,
		r.RoleID,
	)

	if err != nil {
		return err
	}

	log.WithFields(logrus.Fields{
		"msg_type": "add-user-role",
		"status":   "successful",
		"user":     u.Username,
		"role":     r.Rolename,
	}).Info("Add user to role.")

	return nil
}

func (u *MembershipUser) RemoveFromRole(role string) error {
	u.Lock()
	defer u.Unlock()

	var r MembershipRole
	err := r.GetRoleByName(role)

	if err != nil {
		return err
	}

	found, err := r.HasMemberID(u.UserID)

	if err != nil {
		return err
	}

	if !found {
		return nil
	}

	tx, err := db.Begin()

	if err != nil {
		return err
	}

	defer tx.Rollback()

	query := `
		UPDATE wmeter.user_role
		   SET valid_until = current_timestamp
		 WHERE user_id = $1
		   AND role_id = $2
	`

	_, err = tx.Exec(
		query,
		u.UserID,
		r.RoleID,
	)

	if err != nil {
		return err
	}

	query = `
		INSERT INTO wmeter.user_role_history
		SELECT *
		  FROM wmeter.user_role
		 WHERE user_id = $1
		   AND role_id = $2
	`

	_, err = tx.Exec(
		query,
		u.UserID,
		r.RoleID,
	)

	if err != nil {
		return err
	}

	query = `
		DELETE FROM wmeter.user_role
		 WHERE user_id = $1
		   AND role_id = $2
	`

	_, err = tx.Exec(
		query,
		u.UserID,
		r.RoleID,
	)

	if err != nil {
		return err
	}

	tx.Commit()

	log.WithFields(logrus.Fields{
		"msg_type": "remove-user-role",
		"status":   "successful",
		"user":     u.Username,
		"role":     r.Rolename,
	}).Info("Remove user from role.")

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

	hashedPassword, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	password := base64.StdEncoding.EncodeToString(hashedPassword)

	query := `
		SELECT password_id,
		       password,
			   password_salt,
			   valid_until
		  FROM wmeter.user_password
		 WHERE user_id = $1
	`

	err = tx.QueryRow(query, u.UserID).Scan(
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

	hashBytes, err := base64.StdEncoding.DecodeString(hashedPassword)

	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword(hashBytes, passBytes)

	if err != nil {
		return false, err
	}

	return true, nil
}
