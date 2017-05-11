package main

func initializeDatabase() error {
	var err error

	audit.Log(false, nil, "initialize", "adding rquests")
	err = addRequests()

	audit.Log(false, nil, "initialize", "adding roles")
	err = addRoles()

	audit.Log(false, nil, "initialize", "adding system params")
	err = addSystemParams()

	return err
}

type urlRequest struct {
	request_title string
	request_template string
	request_url string
	controller string
	action string
	redirect_url string
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

func addRequests() error {
	var _found bool

	requests := []urlRequest {
		// pages
		urlRequest {"Index", "home/index.html", "index", 	"Home", "Index", 	"-", "-" },
		urlRequest {"About", "home/about.html", "about", 	"Home", "-", 	"-", "-" },
		urlRequest {"Login", "home/login.html", "login", 	"Home", "-", 	"-", "-" },
		urlRequest {"Register", "home/register.html", "register", 	"Home", "-", 	"-", "-" },
		urlRequest {"Change Password", "home/change-password.html", "change-password", 	"Home", "-", 	"-", "-" },
		// gets
		urlRequest {"Logout", "-", "logout", 	"Home", "Logout", 	"/", "-" },
		urlRequest {"Exchange Rates", "-", "exchange-rates", 	"Home", "GetExchangeRates", 	"-", "-" },
		// posts
		urlRequest {"Login", "-", "perform-login", 	"Home", "Login",  "index", "login" },
		urlRequest {"Logout", "-", "perform-logout", "Home", "Logout", "login", "login" },
		urlRequest {"Register", "-", "perform-register", "Home", "Register", "login", "register" },
		urlRequest {"Change Password", "-", "perform-change-password", "Home", "ChangePassword", "change-password", "change-password" },
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

	for _, req := range requests {
		err := db.QueryRow(queryExists, req.request_url).Scan(&_found)

		if err != nil {
			return err
		}

		if !_found {
			_, err = db.Exec(
				queryAdd,
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

	return  nil
}

func addRoles() error {
	var _found bool

	roles := []userRole {
		userRole { "Administrator" },
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

	for _, r := range roles {
		err := db.QueryRow(queryExists, r.role).Scan(&_found)

		if err != nil {
			return err
		}

		if !_found {
			_, err = db.Exec(
				queryAdd,
				r.role,
			)

			if err != nil {
				return err
			}
		}
	}

	return  nil
}

func addSystemParams() error {
	var _found bool

	params := []systemParams {
		systemParams { "password-rules", "change-interval", "30" },
		systemParams { "password-rules", "password-fail-interval", "10" },
		systemParams { "password-rules", "max-allowed-failed-atmpts", "3" },
		systemParams { "password-rules", "not-repeat-last-x-passwords", "5" },
		systemParams { "password-rules", "min-characters", "8" },
		systemParams { "password-rules", "min-letters", "2" },
		systemParams { "password-rules", "min-capitals", "1" },
		systemParams { "password-rules", "min-digits", "1" },
		systemParams { "password-rules", "min-non-alpha-numerics", "1" },
		systemParams { "password-rules", "allow-repetitive-characters", "0" },
		systemParams { "password-rules", "can-contain-username", "0" },
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

	for _, p := range params {
		err := db.QueryRow(queryExists, p.param_group, p.param).Scan(&_found)

		if err != nil {
			return err
		}

		if !_found {
			_, err = db.Exec(
				queryAdd,
				p.param_group,
				p.param,
				p.val,
			)

			if err != nil {
				return err
			}
		}
	}

	return  nil
}
