package main

import "database/sql"

func initializeDatabase() error {
	var err error

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	audit.Log(false, nil, "initialize", "rquests")
	err = addRequests(tx)

	audit.Log(false, nil, "initialize", "roles")
	err = addRoles(tx)

	audit.Log(false, nil, "initialize", "system params")
	err = addSystemParams(tx)

	tx.Commit()

	return err
}

type urlRequest struct {
	request_title     string
	request_template  string
	request_url       string
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
	var _found bool

	requests := []urlRequest{
		// pages
		{"Index", "home/index.html", "index", "Home", "Index", "-", "-"},
		{"About", "home/about.html", "about", "Home", "-", "-", "-"},
		{"Login", "home/login.html", "login", "Home", "-", "-", "-"},
		{"Register", "home/register.html", "register", "Home", "-", "-", "-"},
		{"Change Password", "home/change-password.html", "change-password", "Home", "-", "-", "-"},
		// gets
		{"Logout", "-", "logout", "Home", "Logout", "/", "-"},
		{"Exchange Rates", "-", "exchange-rates", "Home", "GetExchangeRates", "-", "-"},
		// posts
		{"Login", "-", "perform-login", "Home", "Login", "index", "login"},
		{"Logout", "-", "perform-logout", "Home", "Logout", "login", "login"},
		{"Register", "-", "perform-register", "Home", "Register", "login", "register"},
		{"Change Password", "-", "perform-change-password", "Home", "ChangePassword", "change-password", "change-password"},
	}

	queryExists := dbUtils.PQuery(`
		select exists(
		    select 1
		      from request
		    where request_url = ?
		)
	`)

	queryAdd := dbUtils.PQuery(`
		insert into request (
			request_title,
			request_template,
			request_url,
			controller,
			action,
			redirect_url,
			redirect_on_error
		)
		values (?, ?, ?, ?, ?, ?, ?)
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
		err := stmtE.QueryRow(req.request_url).Scan(&_found)

		if err != nil {
			return err
		}

		if !_found {
			_, err = stmtAdd.Exec(
				req.request_title,
				req.request_template,
				req.request_url,
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
	var _found bool

	roles := []userRole{
		{"Administrator"},
	}

	queryExists := dbUtils.PQuery(`
		select exists(
		    select 1
		      from role
		    where lower(role) = lower(?)
		)
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
		err := stmtE.QueryRow(r.role).Scan(&_found)

		if err != nil {
			return err
		}

		if !_found {
			_, err = stmtAdd.Exec(r.role)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func addSystemParams(tx *sql.Tx) error {
	var _found bool

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
		select exists(
		    select 1
		      from system_params
		    where param_group = ?
		      and param = ?
		)
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
		err := stmtE.QueryRow(p.param_group, p.param).Scan(&_found)

		if err != nil {
			return err
		}

		if !_found {
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
