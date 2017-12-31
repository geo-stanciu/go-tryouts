package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
	"unicode"

	"strings"

	"github.com/geo-stanciu/go-utils/utils"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	// ValidationFailed - validation failed
	ValidationFailed int = 0
	// ValidationOK - validation OK
	ValidationOK int = 1
	// ValidationTemporaryPassword - validation temporary password
	ValidationTemporaryPassword int = 2
)

// MembershipUser - membership user helper
type MembershipUser struct {
	sync.RWMutex
	UserID   int
	Username string
	Name     string
	Surname  string
	Email    string
	Password string `json:"-"`
}

// UserExists - user exists
func (u *MembershipUser) UserExists(user string) (bool, error) {
	u.RLock()
	defer u.RUnlock()

	found := false

	query := dbUtils.PQuery(`
	    SELECT CASE WHEN EXISTS (
	        SELECT 1
	          FROM "user"
	         WHERE loweredusername = lower(?)
	    ) THEN 1 ELSE 0 END
	    FROM dual
	`)

	err := db.QueryRow(query, user).Scan(&found)

	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, err
	}

	return found, nil
}

// GetUserByName - get user by name
func (u *MembershipUser) GetUserByName(user string) error {
	u.Lock()
	defer u.Unlock()

	query := dbUtils.PQuery(`
	    SELECT user_id,
	           username,
	           name,
	           surname,
	           email
	      FROM "user"
	     WHERE loweredusername = lower(?)
	`)

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

// GetUserByID - get user by id
func (u *MembershipUser) GetUserByID(userID int) error {
	u.Lock()
	defer u.Unlock()

	query := dbUtils.PQuery(`
	    SELECT user_id,
	        username,
	        name,
	        surname,
	        email
	     FROM "user"
	    WHERE user_id = ?
	`)

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

	query := dbUtils.PQuery(`
	    SELECT CASE WHEN EXISTS (
	        SELECT 1
	          FROM "user"
	         WHERE loweredusername = LOWER(?)
	           AND user_id <> ?
	    ) THEN 1 ELSE 0 END
	    FROM dual
	`)

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

// Save - save user details
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

	dt := time.Now().UTC()

	if u.UserID <= 0 {
		query := dbUtils.PQuery(`
		    INSERT INTO "user" (
		        username,
		        loweredusername,
		        name,
		        surname,
		        email,
		        loweredemail,
		        creation_time,
		        last_update
		    )
		    VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`)

		_, err = tx.Exec(
			query,
			u.Username,
			strings.ToLower(u.Username),
			u.Name,
			u.Surname,
			u.Email,
			strings.ToLower(u.Email),
			dt,
			dt,
		)

		if err != nil {
			return err
		}

		query = dbUtils.PQuery(`
		    SELECT user_id FROM "user" WHERE loweredusername = ?
		`)

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

		err = u.changePassword(tx)
		if err != nil {
			return err
		}

		audit.Log(nil, "add-user", "Add new user.", "new", u)
	} else {
		var old MembershipUser
		err = old.GetUserByID(u.UserID)
		if err != nil {
			return err
		}

		dt := time.Now().UTC()

		if !u.Equals(&old) {
			query = dbUtils.PQuery(`
			    UPDATE "user"
			       SET username     = ?,
			           loweredusername = ?,
			           name            = ?,
			           surname         = ?,
			           email           = ?,
			           loweredemail    = ?,
			           last_update     = ?
			     WHERE user_id = ?
			`)

			_, err = tx.Exec(
				query,
				u.Username,
				strings.ToLower(u.Username),
				u.Name,
				u.Surname,
				u.Email,
				strings.ToLower(u.Email),
				dt,
				u.UserID,
			)

			if err != nil {
				return err
			}

			audit.Log(nil, "update-user", "Update user.", "old", old, "new", u)
		}

		if len(u.Password) > 0 {
			err = u.changePassword(tx)
			if err != nil {
				return err
			}
		}
	}

	tx.Commit()

	return nil
}

// Activate - activates the user
func (u *MembershipUser) Activate() error {
	query := dbUtils.PQuery(`
		UPDATE "user"
		   SET activated       = ?,
			   activation_time = ?
		 WHERE user_id = ?
	       AND activated = ?
	`)

	dt := time.Now().UTC()

	_, err := db.Exec(
		query,
		1,
		dt,
		u.UserID,
		0,
	)

	if err != nil {
		return err
	}

	return nil
}

// GetUserRoles - get user roles
func (u *MembershipUser) GetUserRoles() ([]*MembershipRole, error) {
	u.RLock()
	defer u.RUnlock()

	var roles []*MembershipRole

	dt := time.Now().UTC()

	query := dbUtils.PQuery(`
	    SELECT r.role_id,
	           r.role
	      FROM user_role ur
	      JOIN role r ON (ur.role_id = r.role_id)
	     WHERE ur.user_id = ?
	       AND ur.valid_from <= ?
	       AND (ur.valid_until is null OR ur.valid_until > ?)
	     ORDER BY r.role
	`)

	var err error
	err = dbUtils.ForEachRow(query, func(row *sql.Rows) {
		var r MembershipRole
		err = row.Scan(&r.RoleID, &r.Rolename)
		if err != nil {
			return
		}

		roles = append(roles, &r)
	}, u.UserID,
		dt,
		dt)

	if err != nil {
		return nil, err
	}

	return roles, nil
}

// AddToRole - add user to role
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

	dt := time.Now().UTC()

	query := dbUtils.PQuery(`
	    INSERT INTO user_role (
	        user_id,
	        role_id,
	        valid_from
	    )
	    VALUES (?, ?, ?)
	`)

	_, err = db.Exec(
		query,
		u.UserID,
		r.RoleID,
		dt,
	)

	if err != nil {
		return err
	}

	audit.Log(nil, "add-user-role", "Add user to role.", "user", u.Username, "role", r.Rolename)

	return nil
}

// RemoveFromRole - remove user from role
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

	dt := time.Now().UTC()

	query := dbUtils.PQuery(`
	    UPDATE user_role
	       SET valid_until = ?
	     WHERE user_id = ?
	       AND role_id = ?
	`)

	_, err = tx.Exec(
		query,
		dt,
		u.UserID,
		r.RoleID,
	)

	if err != nil {
		return err
	}

	query = dbUtils.PQuery(`
	    INSERT INTO user_role_history
	    SELECT *
	      FROM user_role
	     WHERE user_id = ?
	       AND role_id = ?
	`)

	_, err = tx.Exec(
		query,
		u.UserID,
		r.RoleID,
	)

	if err != nil {
		return err
	}

	query = dbUtils.PQuery(`
	    DELETE FROM user_role
	     WHERE user_id = ?
	       AND role_id = ?
	`)

	_, err = tx.Exec(
		query,
		u.UserID,
		r.RoleID,
	)

	if err != nil {
		return err
	}

	tx.Commit()

	audit.Log(nil, "remove-user-role", "Remove user from role.", "user", u.Username, "role", r.Rolename)

	return nil
}

func (u *MembershipUser) passwordAlreadyUsed(tx *sql.Tx, params *SystemParams) (bool, int, error) {
	notRepeatPasswords := params.GetInt(NotRepeatLastXPasswords)

	if notRepeatPasswords <= 0 {
		return false, notRepeatPasswords, nil
	}

	var hashedPassword string
	var passwordSalt string

	query := dbUtils.PQuery(`
	    SELECT CASE
	             WHEN password is null THEN
	               '-'
	             ELSE
	               password
	           END AS password,
	           CASE
	             WHEN password_salt is null THEN
	               '-'
	             ELSE
	               password_salt
	           END AS password_salt
	      FROM user_password
	     WHERE user_id = ?
	     ORDER BY password_id DESC
	     LIMIT ?
	`)

	var err error
	err = dbUtils.ForEachRow(query, func(row *sql.Rows) {
		err = row.Scan(&hashedPassword, &passwordSalt)
		if err != nil {
			return
		}

		passBytes := []byte(passwordSalt + u.Password)
		hashBytes, err := base64.StdEncoding.DecodeString(hashedPassword)
		if err != nil {
			return
		}

		err = bcrypt.CompareHashAndPassword(hashBytes, passBytes)
		if err == nil {
			return
		}
	}, u.UserID,
		notRepeatPasswords)

	if err != nil {
		return true, notRepeatPasswords, err
	}

	return false, notRepeatPasswords, nil
}

func (u *MembershipUser) changePassword(tx *sql.Tx) error {
	params := SystemParams{}
	err := params.LoadByGroup(PasswordRules)
	if err != nil {
		return err
	}

	alreadyUsed, notRepeatPasswords, err := u.passwordAlreadyUsed(tx, &params)
	if err != nil {
		return err
	}

	if alreadyUsed {
		return fmt.Errorf("Password already used. Can't use the last %d passwords", notRepeatPasswords)
	}

	changeInterval := params.GetInt(ChangeInterval)
	minCharacters := params.GetInt(MinCharacters)
	minLetters := params.GetInt(MinLetters)
	minCapitals := params.GetInt(MinCapitals)
	minDigits := params.GetInt(MinDigits)
	minNonAlphaNumerics := params.GetInt(MinNonAlphaNumerics)
	allowRepetitiveCharacters := params.GetInt(AllowRepetitiveCharacters)
	canContainUsername := params.GetInt(CanContainUsername)

	if minCharacters > 0 && len(u.Password) < minCharacters {
		return fmt.Errorf("Password must have at least %d characters", minCharacters)
	}

	letters := 0
	capitals := 0
	digits := 0
	nonalphanumerics := 0

	for _, c := range u.Password {
		if c >= 65 && c <= 90 {
			letters++
			capitals++
		} else if c >= 97 && c <= 122 {
			letters++
		} else if unicode.IsNumber(c) {
			digits++
		} else {
			nonalphanumerics++
		}
	}

	if minLetters > 0 && letters < minLetters {
		return fmt.Errorf("Password must contain at least %d letter(s)", minLetters)
	}

	if minCapitals > 0 && capitals < minCapitals {
		return fmt.Errorf("Password must contain at least %d capital letter(s)", minCapitals)
	}

	if minDigits > 0 && digits < minDigits {
		return fmt.Errorf("Password must contain at least %d digit(s)", minDigits)
	}

	if minNonAlphaNumerics > 0 && nonalphanumerics < minNonAlphaNumerics {
		return fmt.Errorf("Password must contain at least %d non alpha-numeric character(s)", minNonAlphaNumerics)
	}

	if allowRepetitiveCharacters <= 0 && utils.ContainsRepeatingGroups(u.Password) {
		return fmt.Errorf("Password must not contain repetitive groups of characters")
	}

	if canContainUsername <= 0 {
		lowerUsername := strings.ToLower(u.Username)
		lowerPass := strings.ToLower(u.Password)

		if strings.Contains(lowerPass, lowerUsername) {
			return fmt.Errorf("Password must not contain the username")
		}
	}

	saltBytes := uuid.NewV4()
	salt := saltBytes.String()

	passwordBytes := []byte(salt + u.Password)
	hashedPassword, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	password := base64.StdEncoding.EncodeToString(hashedPassword)

	dt := time.Now().UTC()

	query := dbUtils.PQuery(`
	    UPDATE user_password
	       SET valid_until = ?
	     WHERE user_id = ?
	       AND valid_from <= ?
	       AND (valid_until is null OR valid_until > ?)
	`)

	_, err = tx.Exec(query, dt, u.UserID, dt, dt)
	if err != nil {
		return err
	}

	until := dt.Add(time.Duration(changeInterval*24) * time.Hour)

	if changeInterval > 0 {
		query = dbUtils.PQuery(`
			INSERT INTO user_password (
				user_id,
				password,
				password_salt,
				valid_from,
				valid_until
			)
			VALUES(?, ?, ?, ?, ?)
		`)

		_, err = tx.Exec(
			query,
			u.UserID,
			password,
			salt,
			dt,
			until,
		)
	} else {
		query = dbUtils.PQuery(`
			INSERT INTO user_password (
				user_id,
				password,
				password_salt,
				valid_from
			)
			VALUES(?, ?, ?, ?)
		`)

		_, err = tx.Exec(
			query,
			u.UserID,
			password,
			salt,
			dt,
		)
	}

	if err != nil {
		return err
	}

	query = dbUtils.PQuery(`
	    UPDATE "user"
	       SET last_password_change = ?
	     WHERE user_id = ?
	`)

	_, err = tx.Exec(query, dt, u.UserID)
	if err != nil {
		return err
	}

	return nil
}

// Equals - test equality between User structs
func (u *MembershipUser) Equals(usr *MembershipUser) bool {
	if u == nil && usr != nil ||
		u != nil && usr == nil ||
		u.UserID != usr.UserID ||
		u.Username != usr.Username ||
		u.Name != usr.Name ||
		u.Surname != usr.Surname ||
		u.Email != usr.Email {

		return false
	}

	return true
}

type validateUserUtil struct {
	UserID         int    `sql:"user_id"`
	HashedPassword string `sql:"hashed_password"`
	PasswordSalt   string `sql:"password_salt"`
	Activated      int    `sql:"activated"`
	LockedOut      int    `sql:"locked_out"`
	Valid          int    `sql:"valid"`
	Temporary      int    `sql:"temporary"`
}

// ValidateUserPassword - check user and password validity
func ValidateUserPassword(user string, pass string, ip string) (int, error) {
	dt := time.Now().UTC()

	query := dbUtils.PQuery(`
	    SELECT u.user_id,
	           CASE
	             WHEN p.password is null THEN
	               '-'
	             ELSE
	               p.password
	           END AS hashed_password,
	           CASE
	             WHEN p.password_salt is null THEN
	               '-'
	             ELSE
	               p.password_salt
	           END AS password_salt,
	           activated,
	           locked_out,
	           valid,
	           p.temporary
	      FROM "user" u
	      LEFT OUTER JOIN user_password p ON (u.user_id = p.user_id)
	     WHERE loweredusername = lower(?)
	       AND p.valid_from <= ?
	       AND (p.valid_until is null OR p.valid_until > ?)
	`)

	testUser := validateUserUtil{}
	err := dbUtils.RunQuery(query, &testUser, user, dt, dt)

	switch {
	case err == sql.ErrNoRows:
		return ValidationFailed, fmt.Errorf("username \"%s\" not found or password expired", user)
	case err != nil:
		return ValidationFailed, err
	}

	if testUser.LockedOut > 0 {
		return ValidationFailed, fmt.Errorf("username \"%s\" is locked out", user)
	}

	if testUser.Activated <= 0 {
		return ValidationFailed, fmt.Errorf("username \"%s\" is not activated", user)
	}

	if testUser.Valid <= 0 {
		return ValidationFailed, fmt.Errorf("username \"%s\" is not valid", user)
	}

	hasIPs := false
	foundIP := false

	query = dbUtils.PQuery(`
		SELECT ip FROM user_ip WHERE user_id = ?
	`)

	err = dbUtils.ForEachRow(query, func(row *sql.Rows) {
		hasIPs = true
		var addr string
		err = row.Scan(&addr)
		if err != nil {
			return
		}

		if addr == ip {
			foundIP = true
		}
	}, testUser.UserID)

	if err != nil {
		return ValidationFailed, err
	}

	if hasIPs && !foundIP {
		return ValidationFailed, fmt.Errorf("IP not accepted for \"%s\"", user)
	}

	passBytes := []byte(testUser.PasswordSalt + pass)
	hashBytes, err := base64.StdEncoding.DecodeString(testUser.HashedPassword)
	if err != nil {
		return ValidationFailed, err
	}

	err = bcrypt.CompareHashAndPassword(hashBytes, passBytes)
	if err != nil {
		failedUserPasswordValidation(testUser.UserID, user)
		return ValidationFailed, err
	}

	if testUser.Temporary > 0 {
		return ValidationTemporaryPassword, nil
	}

	return ValidationOK, nil
}

var passFailLock sync.Mutex

func failedUserPasswordValidation(userID int, user string) {
	passFailLock.Lock()
	defer passFailLock.Unlock()

	var failedPasswordAtempts int
	var firstFailedPassword time.Time
	var maxAllowedFailedAtmpts int
	var passwordFailInterval int
	var passwordStartInterval time.Time
	newFail := 0

	params := SystemParams{}
	err := params.LoadByGroup(PasswordRules)
	if err != nil {
		audit.Log(err, "failed-login", "Operation error.", "user", user)
		return
	}

	passwordFailInterval = params.GetInt(PasswordFailInterval)
	maxAllowedFailedAtmpts = params.GetInt(MaxAllowedFailedAtmpts)

	tx, err := db.Begin()
	if err != nil {
		audit.Log(err, "failed-login", "Operation error.", "user", user)
		return
	}
	defer tx.Rollback()

	query := dbUtils.PQuery(`
	    SELECT failed_password_atmpts,
	           CASE
	             WHEN first_failed_password is null then
	               TIMESTAMP ?
	             ELSE
	               first_failed_password
	            END AS first_failed_password
	      FROM "user" u
	     WHERE user_id = ?
	`)

	err = tx.QueryRow(query, "1970-01-01 00:00:00", userID).Scan(
		&failedPasswordAtempts,
		&firstFailedPassword,
	)

	switch {
	case err == sql.ErrNoRows:
		err1 := fmt.Sprintf("username \"%s\" not found", user)
		audit.Log(err, "failed-login", err1, "user", user)
		return
	case err != nil:
		err1 := fmt.Sprintf("username \"%s\" not found", user)
		audit.Log(err, "failed-login", err1, "user", user)
		return
	}

	passwordStartInterval = time.Now().UTC().Add(time.Duration(-1*passwordFailInterval) * time.Minute)

	if firstFailedPassword.Before(passwordStartInterval) {
		newFail = 1
	}

	dt := time.Now().UTC()

	query = dbUtils.PQuery(`
	    UPDATE "user"
	       SET failed_password_atmpts = CASE WHEN ? = 1 THEN
	                                        1
	                                    ELSE
	                                        failed_password_atmpts + 1
	                                    END,
	           first_failed_password  = CASE WHEN ? = 1 THEN
	                                       ?
	                                    ELSE
	                                        first_failed_password
	                                    END,
	           last_failed_password   = ?
	     WHERE user_id = ?
	`)

	_, err = tx.Exec(query, newFail, newFail, dt, dt, userID)

	if err != nil {
		audit.Log(err, "failed-login", "Failed to setup failed password params.", "user", user)
		return
	}

	query = dbUtils.PQuery(`
	    SELECT failed_password_atmpts
	      FROM "user" u
	     WHERE user_id = ?
	`)

	err = tx.QueryRow(query, userID).Scan(
		&failedPasswordAtempts,
	)

	switch {
	case err == sql.ErrNoRows:
		err1 := fmt.Sprintf("username \"%s\" not found", user)
		audit.Log(err, "failed-login", err1, "user", user)
		return
	case err != nil:
		err1 := fmt.Sprintf("username \"%s\" not found", user)
		audit.Log(err, "failed-login", err1, "user", user)
		return
	}

	if failedPasswordAtempts >= maxAllowedFailedAtmpts {
		query = dbUtils.PQuery(`
		    UPDATE "user"
		       SET locked_out = 1
		     WHERE user_id = ?
		`)

		_, err = tx.Exec(query, userID)

		if err != nil {
			audit.Log(err, "failed-login", "User locked out.", "user", user)
			// return // commented on purpose - Geo 18.03.2017
		}

		dt := time.Now().UTC()

		query = dbUtils.PQuery(`
		    UPDATE user_password
		       SET valid_until = ?
		     WHERE user_id = ?
		       AND valid_from <= ?
		       AND (valid_until is null OR valid_until > ?)
		`)

		_, err = tx.Exec(query, dt, userID, dt, dt)
		if err != nil {
			audit.Log(err, "failed-login", "Failed to invalidate user password.", "user", user)
			// return // commented on purpose - Geo 17.03.2017
		}

		msg := "User password invalidated for multiple failed attempts"

		audit.Log(err, "failed-login", msg, "user", user)
	}

	tx.Commit()

	audit.Log(nil, "failed-login", "Wrong password", "user", user)
}
