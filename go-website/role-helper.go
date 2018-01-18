package main

import (
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// MembershipRole - role utils
type MembershipRole struct {
	sync.RWMutex
	RoleID   int    `sql:"role_id"`
	Rolename string `sql:"role"`
}

var membershipRoleLock sync.RWMutex

// RoleExists - role exists
func (r *MembershipRole) RoleExists(role string) (bool, error) {
	found := false

	pq := dbUtils.PQuery(`
	    SELECT CASE WHEN EXISTS (
	        SELECT 1
	          FROM role
	         WHERE lower(role) = lower(?)
	    ) THEN 1 ELSE 0 END
	    FROM dual
	`, role)

	err := db.QueryRow(pq.Query, pq.Args...).Scan(&found)
	if err != nil {
		return false, err
	}

	return found, nil
}

// GetRoleByName - get role by name
func (r *MembershipRole) GetRoleByName(role string) error {
	r.Lock()
	defer r.Unlock()

	pq := dbUtils.PQuery(`
	    SELECT role_id,
	           role
	      FROM role
	     WHERE lower(role) = lower(?)
	`, role)

	err := dbUtils.RunQuery(pq, r)

	switch {
	case err == sql.ErrNoRows:
		return fmt.Errorf("Role \"%s\" not found", role)
	case err != nil:
		return err
	}

	return nil
}

// GetRoleByID - get role by ID
func (r *MembershipRole) GetRoleByID(roleID int) error {
	r.Lock()
	defer r.Unlock()

	pq := dbUtils.PQuery(`
	    SELECT role_id,
	           role
	      FROM role
	     WHERE role_id = ?
	`, roleID)

	err := dbUtils.RunQuery(pq, r)

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

	pq := dbUtils.PQuery(`
	    SELECT CASE WHEN EXISTS (
	        SELECT 1
	          FROM role
	         WHERE lower(role) = lower(?)
	           AND role_id <> ?
	    ) THEN 1 ELSE 0 END
	    FROM dual
	`, r.Rolename,
		r.RoleID)

	err := tx.QueryRow(pq.Query, pq.Args...).Scan(&found)
	if err != nil {
		return err
	}

	if found {
		return fmt.Errorf("duplicate role \"%s\"", r.Rolename)
	}

	return nil
}

// Save - save role details
func (r *MembershipRole) Save() error {
	membershipRoleLock.Lock()
	defer membershipRoleLock.Unlock()

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
		pq := dbUtils.PQuery(`
		    INSERT INTO role (role) VALUES (?)
		`, r.Rolename)

		_, err = dbUtils.ExecTx(tx, pq)
		if err != nil {
			return err
		}

		pq = dbUtils.PQuery(`
			SELECT role_id FROM role WHERE lower(role) = lower(?)
		`, r.Rolename)

		err = tx.QueryRow(pq.Query, pq.Args...).Scan(&r.RoleID)

		switch {
		case err == sql.ErrNoRows:
			r.RoleID = -1
		case err != nil:
			return err
		}

		audit.Log(nil, "add-role", "Add new role.", "new", r)
	} else {
		var old MembershipRole
		err = old.GetRoleByID(r.RoleID)
		if err != nil {
			return err
		}

		pq := dbUtils.PQuery(`
		    UPDATE role SET role = ? WHERE role_id = ?
		`, r.Rolename,
			r.RoleID)

		_, err = dbUtils.ExecTx(tx, pq)
		if err != nil {
			return err
		}

		audit.Log(nil, "update-role", "Update role.", "old", &old, "new", r)
	}

	tx.Commit()

	return nil
}

// HasMember - role has member
func (r *MembershipRole) HasMember(user string) (bool, error) {
	r.RLock()
	defer r.RUnlock()

	found := false

	dt := time.Now().UTC()

	pq := dbUtils.PQuery(`
	    SELECT CASE WHEN EXISTS (
	        SELECT 1
	          FROM user_role ur
	          JOIN "user" u ON (ur.user_id = u.user_id)
	         WHERE u.loweredusername =  lower(?)
	           AND ur.role_id        =  ?
	           AND ur.valid_from     <= ?
	           AND (ur.valid_until is null OR ur.valid_until > ?)
	    ) THEN 1 ELSE 0 END
	    FROM dual
	`, user,
		r.RoleID,
		dt,
		dt)

	err := db.QueryRow(pq.Query, pq.Args...).Scan(&found)
	if err != nil {
		return false, err
	}

	return found, nil
}

// HasMemberID - has member ID
func (r *MembershipRole) HasMemberID(userID int) (bool, error) {
	r.RLock()
	defer r.RUnlock()

	found := false

	dt := time.Now().UTC()

	pq := dbUtils.PQuery(`
	    SELECT CASE WHEN EXISTS (
	        SELECT 1
	          FROM user_role ur
	         WHERE ur.user_id =  ?
	           AND ur.role_id =  ?
	           AND ur.valid_from <= ?
	           AND (ur.valid_until is null OR ur.valid_until > ?)
	    ) THEN 1 ELSE 0 END
	    FROM dual
	`, userID,
		r.RoleID,
		dt,
		dt)

	err := db.QueryRow(pq.Query, pq.Args...).Scan(&found)
	if err != nil {
		return false, err
	}

	return found, nil
}

// IsUserInRole - Is user in role
func IsUserInRole(user string, role string) (bool, error) {
	found := false

	dt := time.Now().UTC()

	pq := dbUtils.PQuery(`
	    SELECT CASE WHEN EXISTS (
	        SELECT 1
	          FROM user_role ur
	          JOIN "user" u ON (ur.user_id = u.user_id)
	          JOIN role r ON (ur.role_id = r.role_id)
	         WHERE u.loweredusername =  lower(?)
	           AND lower(r.role)     =  lower(?)
	           AND ur.valid_from     <= ?
	           AND (ur.valid_until is null OR ur.valid_until > ?)
	    ) THEN 1 ELSE 0 END
	    FROM dual
	`, user,
		role,
		dt,
		dt)

	err := db.QueryRow(pq.Query, pq.Args...).Scan(&found)
	if err != nil {
		return false, err
	}

	return found, nil
}
