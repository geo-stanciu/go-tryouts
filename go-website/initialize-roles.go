package main

import "database/sql"

type userRole struct {
	role string
}

func addRoles(tx *sql.Tx) error {
	roles := []userRole{
		{"Administrator"},
		{"Member"},
		{"All"},
	}

	for _, r := range roles {
		mrole := MembershipRole{tx: tx}
		mrole.Rolename = r.role

		found, err := mrole.Exists()
		if err != nil {
			return err
		}

		if found {
			continue
		}

		err = mrole.Save()
		if err != nil {
			return err
		}
	}

	return nil
}
