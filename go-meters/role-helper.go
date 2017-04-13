package main

import (
	"database/sql"
	"fmt"
	"sync"
)

type MembershipRole struct {
	sync.RWMutex
	RoleID   int
	Rolename string
}

func (r *MembershipRole) RoleExists(role string) (bool, error) {
	r.RLock()
	defer r.RUnlock()

	found := false

	query := `
		SELECT EXISTS(
			SELECT 1
			  FROM wmeter.role
			 WHERE lower(role) = lower($1)
		)
	`

	err := db.QueryRow(query, role).Scan(&found)

	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, err
	}

	return found, nil
}

func (r *MembershipRole) GetRoleByName(role string) error {
	r.Lock()
	defer r.Unlock()

	query := `
        SELECT role_id,
		       role
          FROM wmeter.role
         WHERE lower(role) = lower($1)
    `

	err := db.QueryRow(query, role).Scan(
		&r.RoleID,
		&r.Rolename)

	switch {
	case err == sql.ErrNoRows:
		return fmt.Errorf("Role \"%s\" not found", role)
	case err != nil:
		return err
	}

	return nil
}

func (r *MembershipRole) GetRoleByID(roleID int) error {
	r.Lock()
	defer r.Unlock()

	query := `
        SELECT role_id,
		       role
          FROM wmeter.role
         WHERE role_id = $1
    `

	err := db.QueryRow(query, roleID).Scan(
		&r.RoleID,
		&r.Rolename)

	switch {
	case err == sql.ErrNoRows:
		return fmt.Errorf("role not found")
	case err != nil:
		return err
	}

	return nil
}

func (r *MembershipRole) testSaveRole(tx *sql.Tx) error {
	if len(r.Rolename) == 0 {
		return fmt.Errorf("unknown role \"%s\"", r.Rolename)
	}

	var found bool

	query := `
        SELECT EXISTS(
			SELECT 1
		      FROM wmeter.role
			 WHERE lower(role) = lower($1)
			   AND role_id <> $2
		)
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	err = stmt.QueryRow(r.Rolename, r.RoleID).Scan(&found)

	switch {
	case err == sql.ErrNoRows:
		found = false
	case err != nil:
		return err
	}

	if found {
		return fmt.Errorf("duplicate role \"%s\"", r.Rolename)
	}

	return nil
}

func (r *MembershipRole) Save() error {
	r.Lock()
	defer r.Unlock()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = r.testSaveRole(tx)
	if err != nil {
		return err
	}

	if r.RoleID < 0 {
		query := `
			INSERT INTO wmeter.role (
				role
			)
			VALUES (
				$1
			)
		`

		_, err = tx.Exec(
			query,
			r.Rolename,
		)

		if err != nil {
			return err
		}

		query = `
			SELECT role_id FROM wmeter.role WHERE lower(role) = lower($1)
		`

		err = tx.QueryRow(query, r.Rolename).Scan(&r.RoleID)

		switch {
		case err == sql.ErrNoRows:
			r.RoleID = -1
		case err != nil:
			return err
		}

		Log(false, nil, "add-role", "Add new role.", "new", r)
	} else {
		var old MembershipRole
		err = old.GetRoleByID(r.RoleID)
		if err != nil {
			return err
		}

		query := `
			UPDATE wmeter.role SET role = $1 WHERE role_id = $2
		`

		_, err = tx.Exec(
			query,
			r.Rolename,
			r.Rolename,
		)

		if err != nil {
			return err
		}

		Log(false, nil, "update-role", "Add new role.", "old", old, "new", r)
	}

	tx.Commit()

	return nil
}

func (r *MembershipRole) HasMember(user string) (bool, error) {
	found := false

	query := `
		SELECT EXISTS(
			SELECT 1
			  FROM wmeter.user_role ur
			  JOIN wmeter.user u ON (ur.user_id = u.user_id)
			 WHERE u.loweredusername =  lower($1)
			   AND ur.role_id        =  $2
			   AND ur.valid_from     <= current_timestamp
			   AND (ur.valid_until is null OR ur.valid_until > current_timestamp)
		)
	`

	err := db.QueryRow(query, user, r.RoleID).Scan(&found)

	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, err
	}

	return found, nil
}

func (r *MembershipRole) HasMemberID(userID int) (bool, error) {
	found := false

	query := `
		SELECT EXISTS(
			SELECT 1
			  FROM wmeter.user_role ur
			 WHERE ur.user_id =  $1
			   AND ur.role_id =  $2
			   AND ur.valid_from     <= current_timestamp
			   AND (ur.valid_until is null OR ur.valid_until > current_timestamp)
		)
	`

	err := db.QueryRow(query, userID, r.RoleID).Scan(&found)

	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, err
	}

	return found, nil
}

func IsUserInRole(user string, role string) (bool, error) {
	found := false

	query := `
		SELECT EXISTS(
			SELECT 1
			  FROM wmeter.user_role ur
			  JOIN wmeter.user u ON (ur.user_id = u.user_id)
			  JOIN wmeter.role r ON (ur.role_id = r.role_id)
			 WHERE u.loweredusername =  lower($1)
			   AND lower(r.role)     =  lower($2)
			   AND ur.valid_from     <= current_timestamp
			   AND (ur.valid_until is null OR ur.valid_until > current_timestamp)
		)
	`

	err := db.QueryRow(query, user, role).Scan(&found)

	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, err
	}

	return found, nil
}
