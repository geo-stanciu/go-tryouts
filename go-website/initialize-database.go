package main

import "database/sql"

func initializeDatabase() error {
	var err error

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = addRequests(tx)
	audit.Log(err, "initialize", "requests")

	err = addRoles(tx)
	audit.Log(err, "initialize", "roles")

	err = addMenu(tx)
	audit.Log(err, "initialize", "menus")

	err = addSystemParams(tx)
	audit.Log(err, "initialize", "system params")

	tx.Commit()

	return err
}

type systemParams struct {
	param_group string
	param       string
	val         string
}

func addSystemParams(tx *sql.Tx) error {
	var found bool

	params := []systemParams{
		{"password-rules", "change-interval", "30"},
		{"password-rules", "password-fail-interval", "10"},
		{"password-rules", "max-allowed-failed-atmpts", "3"},
		{"password-rules", "not-repeat-last-x-passwords", "5"},
		{"password-rules", "min-characters", "8"},
		{"password-rules", "min-letters", "2"},
		{"password-rules", "min-capitals", "1"},
		{"password-rules", "min-digits", "1"},
		{"password-rules", "min-non-alpha-numerics", "1"},
		{"password-rules", "allow-repetitive-characters", "0"},
		{"password-rules", "can-contain-username", "0"},
	}

	pqExists := dbUtils.PQuery(`
	    select CASE WHEN EXISTS (
	        select 1
	          from system_params
	         where param_group = ?
	           and param = ?
	    ) THEN 1 ELSE 0 END
	    FROM dual
	`)

	pqAdd := dbUtils.PQuery(`
	   insert into system_params (
	       param_group,
	       param,
	       val
	    )
	    values (?, ?, ?)
	`)

	stmtE, err := tx.Prepare(pqExists.Query)
	if err != nil {
		return err
	}
	defer stmtE.Close()

	stmtAdd, err := tx.Prepare(pqAdd.Query)
	if err != nil {
		return err
	}
	defer stmtAdd.Close()

	for _, p := range params {
		err := stmtE.QueryRow(p.param_group, p.param).Scan(&found)

		if err != nil {
			return err
		}

		if !found {
			_, err = stmtAdd.Exec(
				p.param_group,
				p.param,
				p.val,
			)

			if err != nil {
				return err
			}
		}
	}

	return nil
}
