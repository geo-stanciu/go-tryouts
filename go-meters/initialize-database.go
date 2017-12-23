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

	err = addSystemParams(tx)
	audit.Log(err, "initialize", "system params")

	tx.Commit()

	return err
}

type urlRequest struct {
	request_title     string
	request_template  string
	request_url       string
	request_type      string
	controller        string
	action            string
	redirect_url      string
	redirect_on_error string
}

type userRole struct {
	role string
}

type systemParams struct {
	param_group string
	param       string
	val         string
}

func addRequests(tx *sql.Tx) error {
	var found bool

	requests := []urlRequest{
		// pages
		{"Index", "home/index.html", "index", "GET", "Home", "Index", "-", "-"},
		{"About", "home/about.html", "about", "GET", "Home", "-", "-", "-"},
		{"Login", "home/login.html", "login", "GET", "Home", "-", "-", "-"},
		{"Register", "home/register.html", "register", "GET", "Home", "-", "-", "-"},
		{"Change Password", "home/change-password.html", "change-password", "GET", "Home", "-", "-", "-"},
		// gets
		{"Logout", "-", "logout", "GET", "Home", "Logout", "/", "-"},
		{"Exchange Rates", "-", "exchange-rates", "GET", "Home", "GetExchangeRates", "-", "-"},
		// posts
		{"Login", "-", "login", "POST", "Home", "Login", "index", "login"},
		{"Logout", "-", "logout", "POST", "Home", "Logout", "login", "login"},
		{"Register", "-", "register", "POST", "Home", "Register", "login", "register"},
		{"Change Password", "-", "change-password", "POST", "Home", "ChangePassword", "change-password", "change-password"},
		{"Exchange Rates", "-", "exchange-rates", "POST", "Home", "GetExchangeRates", "-", "-"},
	}

	queryExists := dbUtils.PQuery(`
		select CASE WHEN EXISTS (
		    select 1
		      from request
			where request_url = ?
			  and request_type = ?
		) THEN 1 ELSE 0 END
		FROM dual
	`)

	queryAdd := dbUtils.PQuery(`
		insert into request (
			request_title,
			request_template,
			request_url,
			request_type,
			controller,
			action,
			redirect_url,
			redirect_on_error
		)
		values (?, ?, ?, ?, ?, ?, ?, ?)
	`)

	stmtE, err := tx.Prepare(queryExists)
	if err != nil {
		return err
	}
	defer stmtE.Close()

	stmtAdd, err := tx.Prepare(queryAdd)
	if err != nil {
		return err
	}
	defer stmtAdd.Close()

	for _, req := range requests {
		err := stmtE.QueryRow(req.request_url, req.request_type).Scan(&found)

		if err != nil {
			return err
		}

		if !found {
			_, err = stmtAdd.Exec(
				req.request_title,
				req.request_template,
				req.request_url,
				req.request_type,
				req.controller,
				req.action,
				req.redirect_url,
				req.redirect_on_error,
			)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func addRoles(tx *sql.Tx) error {
	var found bool

	roles := []userRole{
		{"Administrator"},
		{"Member"},
	}

	queryExists := dbUtils.PQuery(`
		select CASE WHEN EXISTS (
		    select 1
		      from role
		    where lower(role) = lower(?)
		) THEN 1 ELSE 0 END
		FROM dual
	`)

	queryAdd := dbUtils.PQuery(`
		insert into role (
			role
		)
		values (?)
	`)

	stmtE, err := tx.Prepare(queryExists)
	if err != nil {
		return err
	}
	defer stmtE.Close()

	stmtAdd, err := tx.Prepare(queryAdd)
	if err != nil {
		return err
	}
	defer stmtAdd.Close()

	for _, r := range roles {
		err := stmtE.QueryRow(r.role).Scan(&found)

		if err != nil {
			return err
		}

		if !found {
			_, err = stmtAdd.Exec(r.role)
			if err != nil {
				return err
			}
		}
	}

	return nil
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

	queryExists := dbUtils.PQuery(`
		select CASE WHEN EXISTS (
		    select 1
		      from system_params
		    where param_group = ?
		      and param = ?
		) THEN 1 ELSE 0 END
		FROM dual
	`)

	queryAdd := dbUtils.PQuery(`
		insert into system_params (
			param_group,
			param,
			val
		)
		values (?, ?, ?)
	`)

	stmtE, err := tx.Prepare(queryExists)
	if err != nil {
		return err
	}
	defer stmtE.Close()

	stmtAdd, err := tx.Prepare(queryAdd)
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
