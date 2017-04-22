package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
	"unicode"

	"strings"

	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type MembershipUser struct {
	sync.RWMutex
	UserID   int
	Username string
	Name     string
	Surname  string
	Email    string
	Password string `json:"-"`
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

		err = u.changePassword(tx)
		if err != nil {
			return err
		}

		Log(false, nil, "add-user", "Add new user.", "new", u)
	} else {
		var old MembershipUser
		err = old.GetUserByID(u.UserID)
		if err != nil {
			return err
		}

		if !u.Equals(&old) {
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

			Log(false, nil, "update-user", "Update user.", "old", old, "new", u)
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
		   AND ur.valid_from     <= current_timestamp
		   AND (ur.valid_until is null OR ur.valid_until > current_timestamp)
		 ORDER BY r.role
	`

	rows, err := db.Query(query, u.UserID)
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

	Log(false, nil, "add-user-role", "Add user to role.", "user", u.Username, "role", r.Rolename)

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

	Log(false, nil, "remove-user-role", "Remove user from role.", "user", u.Username, "role", r.Rolename)

	return nil
}

func (u *MembershipUser) passwordAlreadyUsed(tx *sql.Tx, params *SystemParams) (bool, int, error) {
	notRepeatPasswords := params.GetInt(NotRepeatLastXPasswords)

	if notRepeatPasswords <= 0 {
		return false, notRepeatPasswords, nil
	}

	var hashedPassword string
	var passwordSalt string

	query := `
		SELECT COALESCE(password, '') AS password,
		       COALESCE(password_salt, '') AS password_salt
		  FROM wmeter.user_password
		 WHERE user_id = $1
		 ORDER BY password_id DESC
		 LIMIT $2
	`

	rows, err := tx.Query(query, u.UserID, notRepeatPasswords)
	if err != nil {
		return true, notRepeatPasswords, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&hashedPassword, &passwordSalt)
		if err != nil {
			return true, notRepeatPasswords, err
		}

		passBytes := []byte(passwordSalt + u.Password)
		hashBytes, err := base64.StdEncoding.DecodeString(hashedPassword)
		if err != nil {
			return true, notRepeatPasswords, err
		}

		err = bcrypt.CompareHashAndPassword(hashBytes, passBytes)
		if err == nil {
			return true, notRepeatPasswords, err
		}
	}

	if err := rows.Err(); err != nil {
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

	if allowRepetitiveCharacters <= 0 && containsRepeatingGroups(u.Password) {
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

	query := `
		UPDATE wmeter.user_password p
		   SET valid_until = current_timestamp
		 WHERE user_id = $1
		   AND p.valid_from <= current_timestamp
		   AND (p.valid_until is null OR p.valid_until > current_timestamp)
	`

	_, err = tx.Exec(query, u.UserID)
	if err != nil {
		return err
	}

	query = fmt.Sprintf(`
		INSERT INTO wmeter.user_password (
			user_id,
			password,
			password_salt,
			valid_until
		)
		SELECT $1,
		       $2,
			   $3,
			   CASE WHEN $4 > 0 THEN
			       current_timestamp + interval '%d' day
			   ELSE
			       null
			   END
	`, changeInterval)

	_, err = tx.Exec(
		query,
		u.UserID,
		password,
		salt,
		changeInterval,
	)

	if err != nil {
		return err
	}

	query = `
		UPDATE wmeter.user
		   SET last_password_change = current_timestamp
		 WHERE user_id = $1
	`

	_, err = tx.Exec(query, u.UserID)
	if err != nil {
		return err
	}

	return nil
}

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

func ValidateUserPassword(user string, pass string) (bool, error) {
	var userID int
	var hashedPassword string
	var passwordSalt string
	var activated int
	var lockedOut int
	var valid int

	query := `
        SELECT u.user_id,
		       COALESCE(p.password, '') AS password,
		       COALESCE(p.password_salt, '') AS password_salt,
			   activated,
			   locked_out,
			   valid
          FROM wmeter.user u
          LEFT OUTER JOIN wmeter.user_password p ON (u.user_id = p.user_id)
         WHERE loweredusername = lower($1)
		   AND p.valid_from <= current_timestamp
		   AND (p.valid_until is null OR p.valid_until > current_timestamp)
    `

	err := db.QueryRow(query, user).Scan(
		&userID,
		&hashedPassword,
		&passwordSalt,
		&activated,
		&lockedOut,
		&valid,
	)

	switch {
	case err == sql.ErrNoRows:
		return false, fmt.Errorf("username \"%s\" not found", user)
	case err != nil:
		return false, err
	}

	if lockedOut > 0 {
		return false, fmt.Errorf("username \"%s\" is locked out", user)
	}

	if activated <= 0 {
		return false, fmt.Errorf("username \"%s\" is not activated", user)
	}

	if valid <= 0 {
		return false, fmt.Errorf("username \"%s\" is not valid", user)
	}

	passBytes := []byte(passwordSalt + pass)
	hashBytes, err := base64.StdEncoding.DecodeString(hashedPassword)
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword(hashBytes, passBytes)
	if err != nil {
		failedUserPasswordValidation(userID, user)
		return false, err
	}

	return true, nil
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
		Log(true, err, "failed-login", "Operation error.", "user", user)
		return
	}

	passwordFailInterval = params.GetInt(PasswordFailInterval)
	maxAllowedFailedAtmpts = params.GetInt(MaxAllowedFailedAtmpts)

	tx, err := db.Begin()
	if err != nil {
		Log(true, err, "failed-login", "Operation error.", "user", user)
		return
	}
	defer tx.Rollback()

	query := `
		SELECT failed_password_atmpts,
		       COALESCE(first_failed_password, to_timestamp('1970-01-01', 'yyyy-mm-dd')) AS first_failed_password
		  FROM wmeter.user u
		 WHERE user_id = $1
	`

	err = tx.QueryRow(query, userID).Scan(
		&failedPasswordAtempts,
		&firstFailedPassword,
	)

	switch {
	case err == sql.ErrNoRows:
		err1 := fmt.Sprintf("username \"%s\" not found", user)
		Log(true, err, "failed-login", err1, "user", user)
		return
	case err != nil:
		err1 := fmt.Sprintf("username \"%s\" not found", user)
		Log(true, err, "failed-login", err1, "user", user)
		return
	}

	passwordStartInterval = time.Now().Add(time.Duration(-1*passwordFailInterval) * time.Minute)

	if firstFailedPassword.Before(passwordStartInterval) {
		newFail = 1
	}

	query = `
		UPDATE wmeter.user u
		   SET failed_password_atmpts = CASE WHEN $1 = 1 THEN
		                                    1
										ELSE
										    failed_password_atmpts + 1
		                                END,
		       first_failed_password  = CASE WHEN $2 = 1 THEN
			   			                    current_timestamp
										ELSE
										    first_failed_password
			                            END,
		       last_failed_password   = current_timestamp
		 WHERE user_id = $3
	`

	_, err = tx.Exec(query, newFail, newFail, userID)

	if err != nil {
		Log(true, err, "failed-login", "Failed to setup failed password params.", "user", user)
		return
	}

	query = `
		SELECT failed_password_atmpts
		  FROM wmeter.user u
		 WHERE user_id = $1
	`

	err = tx.QueryRow(query, userID).Scan(
		&failedPasswordAtempts,
	)

	switch {
	case err == sql.ErrNoRows:
		err1 := fmt.Sprintf("username \"%s\" not found", user)
		Log(true, err, "failed-login", err1, "user", user)
		return
	case err != nil:
		err1 := fmt.Sprintf("username \"%s\" not found", user)
		Log(true, err, "failed-login", err1, "user", user)
		return
	}

	if failedPasswordAtempts >= maxAllowedFailedAtmpts {
		query = `
			UPDATE wmeter.user
			   SET locked_out = 1
			 WHERE user_id = $1
		`

		_, err = tx.Exec(query, userID)

		if err != nil {
			Log(true, err, "failed-login", "User locked out.", "user", user)
			// return // commented on purpose - Geo 18.03.2017
		}

		query = `
			UPDATE wmeter.user_password p
			   SET valid_until = current_timestamp
			 WHERE user_id = $1
			   AND valid_from <= current_timestamp
		       AND (valid_until is null OR valid_until > current_timestamp)
		`

		_, err = tx.Exec(query, userID)
		if err != nil {
			Log(true, err, "failed-login", "Failed to invalidate user password.", "user", user)
			// return // commented on purpose - Geo 17.03.2017
		}

		msg := "User password invalidated for multiple failed attempts"

		Log(true, nil, "failed-login", msg, "user", user)
	}

	tx.Commit()

	Log(false, nil, "failed-login", "Wrong password", "user", user)
}
