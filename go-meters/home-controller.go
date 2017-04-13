package main

import (
	"net/http"

	"./models"

	"strings"
)

type HomeController struct {
}

func (HomeController) Index(w http.ResponseWriter, r *http.Request, res *ResponseHelper) (*models.GenericResponseModel, error) {
	return nil, nil
}

func (HomeController) Login(w http.ResponseWriter, r *http.Request, res *ResponseHelper) (*models.GenericResponseModel, error) {
	var lres models.GenericResponseModel

	if res != nil {
		lres.SSuccessURL = res.RedirectURL
		lres.SErrorURL = res.RedirectOnError
	}

	session, _ := getSessionData(r)

	if !session.LoggedIn {
		user := r.FormValue("username")
		pass := r.FormValue("password")

		if len(user) == 0 || len(pass) == 0 {
			lres.BError = true
			lres.SError = "Unknown user or wrong password."
			Log(lres.BError, nil, "login", lres.SError, "user", user)

			return &lres, nil
		}

		success, err := ValidateUserPassword(user, pass)

		if err != nil || !success {
			lres.BError = true
			lres.SError = "Unknown user or wrong password."
			Log(lres.BError, nil, "login", lres.SError, "user", user)

			return &lres, err
		}

		session, err = createSession(w, r, user)

		if err != nil {
			lres.BError = true
			lres.SError = err.Error()
			Log(lres.BError, err, "login", lres.SError, "user", user)

			return &lres, err
		}
	}

	lres.BError = false
	Log(lres.BError, nil, "login", "User logged in.", "user", session.User.Username)
	return &lres, nil
}

func (HomeController) Logout(w http.ResponseWriter, r *http.Request, res *ResponseHelper) (*models.GenericResponseModel, error) {
	var lres models.GenericResponseModel
	var user string

	if res != nil {
		lres.SSuccessURL = res.RedirectURL
		lres.SErrorURL = res.RedirectOnError
	}

	session, _ := getSessionData(r)

	if session.LoggedIn {
		user = session.User.Username
		err := clearSession(w, r)

		if err != nil {
			lres.BError = true
			lres.SError = err.Error()
			Log(lres.BError, err, "logout", lres.SError, "user", user)
			return nil, err
		}
	}

	Log(false, nil, "logout", "User logged out.", "user", user)

	return &lres, nil
}

func (HomeController) Register(w http.ResponseWriter, r *http.Request, res *ResponseHelper) (*models.GenericResponseModel, error) {
	var lres models.GenericResponseModel

	if res != nil {
		lres.SSuccessURL = res.RedirectURL
		lres.SErrorURL = res.RedirectOnError
	}

	user := r.FormValue("username")
	pass := r.FormValue("password")
	confirmPass := r.FormValue("confirm_password")
	name := r.FormValue("name")
	surname := r.FormValue("surname")
	email := r.FormValue("email")

	if len(user) == 0 {
		lres.BError = true
		lres.SError = "User is empty"
		Log(lres.BError, nil, "register", lres.SError, "user", user, "email", email)

		return &lres, nil
	}

	if len(pass) == 0 || pass != confirmPass {
		lres.BError = true
		lres.SError = "Password is empty or is different from it's confirmation."
		Log(lres.BError, nil, "register", lres.SError, "user", user, "email", email)

		return &lres, nil
	}

	if len(email) == 0 {
		lres.BError = true
		lres.SError = "E-mail is empty"
		Log(lres.BError, nil, "register", lres.SError, "user", user, "email", email)

		return &lres, nil
	}

	u := MembershipUser{
		UserID:   -1,
		Username: user,
		Name:     name,
		Surname:  surname,
		Email:    email,
		Password: pass,
	}

	err := u.Save()

	if err != nil {
		lres.BError = true
		lres.SError = err.Error()
		Log(lres.BError, err, "register", lres.SError, "user", user, "email", email)

		return &lres, err
	}

	if isRequestFromLocalhost(r) && strings.ToLower(u.Username) == "admin" {
		err = u.AddToRole("Administrator")

		if err != nil {
			lres.BError = true
			lres.SError = err.Error()
			Log(lres.BError, err, "register", lres.SError, "user", user, "email", email)

			return &lres, err
		}
	}

	lres.BError = false
	lres.SError = "User registered"
	Log(lres.BError, nil, "register", lres.SError, "user", user, "email", email)

	return &lres, nil
}

func (HomeController) ChangePassword(w http.ResponseWriter, r *http.Request, res *ResponseHelper) (*models.GenericResponseModel, error) {
	var lres models.GenericResponseModel

	if res != nil {
		lres.SSuccessURL = res.RedirectURL
		lres.SErrorURL = res.RedirectOnError
	}

	session, _ := getSessionData(r)

	if !session.LoggedIn {
		lres.BError = true
		lres.SError = "User not logged in."
		Log(lres.BError, nil, "change-password", lres.SError, "user", "", "email", "")

		return &lres, nil
	}

	var usr MembershipUser
	err := usr.GetUserByName(session.User.Username)

	if err != nil {
		lres.BError = true
		lres.SError = err.Error()
		Log(lres.BError, err, "change-password", lres.SError, "user", "", "email", "")

		return &lres, nil
	}

	pass := r.FormValue("password")
	newPass := r.FormValue("new_password")
	confirmPass := r.FormValue("confirm_password")

	if len(pass) == 0 {
		lres.BError = true
		lres.SError = "Old password cannot be empty"
		Log(lres.BError, nil, "change-password", lres.SError, "user", usr.Username, "email", usr.Email)

		return &lres, nil
	}

	if len(newPass) == 0 || newPass != confirmPass {
		lres.BError = true
		lres.SError = "Password is empty or is different from it's confirmation."
		Log(lres.BError, nil, "change-password", lres.SError, "user", usr.Username, "email", usr.Email)

		return &lres, nil
	}

	if pass == newPass {
		lres.BError = true
		lres.SError = "The new password must be different from the current one."
		Log(lres.BError, nil, "change-password", lres.SError, "user", usr.Username, "email", usr.Email)

		return &lres, nil
	}

	valid, err := ValidateUserPassword(usr.Username, pass)

	if !valid {
		lres.BError = true
		lres.SError = "Old password is not valid."
		Log(lres.BError, nil, "change-password", lres.SError, "user", usr.Username, "email", usr.Email)

		return &lres, nil
	}

	usr.Password = newPass

	err = usr.Save()

	if err != nil {
		lres.BError = true
		lres.SError = err.Error()
		Log(lres.BError, err, "change-password", lres.SError, "user", usr.Username, "email", usr.Email)

		return &lres, err
	}

	lres.BError = false
	lres.SError = "User password changed"
	Log(lres.BError, nil, "change-password", lres.SError, "user", usr.Username, "email", usr.Email)

	return &lres, nil
}

func (HomeController) GetExchangeRates(w http.ResponseWriter, r *http.Request, res *ResponseHelper) (*models.ExchangeRatesResponseModel, error) {
	var lres models.ExchangeRatesResponseModel

	query := `
		select c.currency, r.exchange_date, r.rate
		  from exchange_rate r
		  join currency c on (r.currency_id = c.currency_id)
		 where exchange_date = (
			 select max(exchange_date) from exchange_rate
		 )
		 order by c.currency, r.exchange_date
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.Rate
		err = rows.Scan(&r.Currency, &r.Date, &r.Value)
		if err != nil {
			return nil, err
		}

		lres.Rates = append(lres.Rates, r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &lres, nil
}
