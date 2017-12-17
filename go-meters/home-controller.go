package main

import (
	"database/sql"
	"net/http"
	"time"

	"./models"

	"strings"

	"fmt"

	"github.com/geo-stanciu/go-utils/utils"
)

type HomeController struct {
}

func (HomeController) Index(w http.ResponseWriter, r *http.Request, res *ResponseHelper) (*models.GenericResponseModel, error) {
	return nil, nil
}

func (HomeController) Login(w http.ResponseWriter, r *http.Request, res *ResponseHelper) (*models.LoginResponseModel, error) {
	var lres models.LoginResponseModel
	var ip string
	var user string
	var pass string
	var err error
	throwErr2Client := true

	if res != nil {
		lres.SSuccessURL = res.RedirectURL
		lres.SErrorURL = res.RedirectOnError
	}

	ip = getClientIP(r)

	sessionData, _ := getSessionData(r)

	if !sessionData.LoggedIn {
		user = r.FormValue("username")
		pass = r.FormValue("password")

		if len(user) == 0 || len(pass) == 0 {
			throwErr2Client = false
			lres, err = loginerr(&lres, err, user, throwErr2Client)
			return &lres, err
		}

		success, err := ValidateUserPassword(user, pass, ip)
		if err != nil || (success != ValidationOK && success != ValidationTemporaryPassword) {
			throwErr2Client = false
			lres, err = loginerr(&lres, err, user, throwErr2Client)
			return &lres, err
		}

		if success == ValidationTemporaryPassword {
			lres.TemporaryPassword = true
		}

		var name string
		var surname string

		query := dbUtils.PQuery(`
			SELECT name, surname
			  FROM user
			 WHERE loweredusername = lower(?)
		`)

		err = db.QueryRow(query, user).Scan(&name, &surname)
		if err != nil {
			lres, err = loginerr(&lres, err, user, throwErr2Client)
			return &lres, err
		}

		sessionData, err = createSession(w, r, user, name, surname, lres.TemporaryPassword)
		if err != nil {
			lres, err = loginerr(&lres, err, user, throwErr2Client)
			return &lres, err
		}

		dt := time.Now().UTC()

		query = dbUtils.PQuery(`
			UPDATE user
			   SET last_connect_time = ?,
			       last_connect_ip   = ?
			 WHERE loweredusername = lower(?)
		`)

		_, err = db.Exec(query, dt, ip, user)
		if err != nil {
			lres, err = loginerr(&lres, err, user, throwErr2Client)
			return &lres, err
		}
	}

	lres.BError = false
	audit.Log(nil, "login", "User logged in.",
		"user", sessionData.User.Username,
		"ip", ip,
		"Temporary Password", lres.TemporaryPassword)

	return &lres, nil
}

func loginerr(lres *models.LoginResponseModel,
	errLogin error,
	user string,
	throwErr2Client bool) (models.LoginResponseModel, error) {

	err := errLogin

	lres.BError = true

	if err != nil && throwErr2Client {
		if err == nil {
			err = fmt.Errorf("Unknown error")
		}

		lres.SError = err.Error()
	} else {
		lres.SError = "Unknown user or wrong password."
	}

	audit.Log(err, "login", lres.SError,
		"user", user,
		"Temporary Password", lres.TemporaryPassword,
	)

	err = fmt.Errorf(lres.SError)
	return *lres, nil
}

func (HomeController) Logout(w http.ResponseWriter, r *http.Request, res *ResponseHelper) (*models.GenericResponseModel, error) {
	var lres models.GenericResponseModel
	var user string

	if res != nil {
		lres.SSuccessURL = res.RedirectURL
		lres.SErrorURL = res.RedirectOnError
	}

	sessionData, _ := getSessionData(r)

	if sessionData.LoggedIn {
		user = sessionData.User.Username
		err := clearSession(w, r)

		if err != nil {
			lres.BError = true
			lres.SError = err.Error()
			audit.Log(err, "logout", lres.SError, "user", user)
			return nil, err
		}
	}

	audit.Log(nil, "logout", "User logged out.", "user", user)

	return &lres, nil
}

func (HomeController) Register(w http.ResponseWriter, r *http.Request, res *ResponseHelper) (*models.GenericResponseModel, error) {
	var lres models.GenericResponseModel
	var err error

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
		err = fmt.Errorf(lres.SError)
		audit.Log(err, "register", lres.SError, "user", user, "email", email)

		return &lres, nil
	}

	if len(pass) == 0 || pass != confirmPass {
		lres.BError = true
		lres.SError = "Password is empty or is different from it's confirmation."
		err = fmt.Errorf(lres.SError)
		audit.Log(err, "register", lres.SError, "user", user, "email", email)

		return &lres, nil
	}

	if len(email) == 0 {
		lres.BError = true
		lres.SError = "E-mail is empty"
		err = fmt.Errorf(lres.SError)
		audit.Log(err, "register", lres.SError, "user", user, "email", email)

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

	err = u.Save()

	if err != nil {
		lres.BError = true
		lres.SError = err.Error()
		audit.Log(err, "register", lres.SError, "user", user, "email", email)

		return &lres, err
	}

	if isRequestFromLocalhost(r) && strings.ToLower(u.Username) == "admin" {
		err = u.AddToRole("Administrator")

		if err != nil {
			lres.BError = true
			lres.SError = err.Error()
			audit.Log(err, "register", lres.SError, "user", user, "email", email)

			return &lres, err
		}
	}

	lres.BError = false
	lres.SError = "User registered"
	audit.Log(nil, "register", lres.SError, "user", user, "email", email)

	return &lres, nil
}

func (HomeController) ChangePassword(w http.ResponseWriter, r *http.Request, res *ResponseHelper) (*models.GenericResponseModel, error) {
	var lres models.GenericResponseModel
	var err error

	if res != nil {
		lres.SSuccessURL = res.RedirectURL
		lres.SErrorURL = res.RedirectOnError
	}

	sessionData, _ := getSessionData(r)

	if !sessionData.LoggedIn {
		lres.BError = true
		lres.SError = "User not logged in."
		err = fmt.Errorf(lres.SError)
		audit.Log(err, "change-password", lres.SError, "user", "", "email", "")

		return &lres, nil
	}

	var usr MembershipUser
	err = usr.GetUserByName(sessionData.User.Username)

	if err != nil {
		lres.BError = true
		lres.SError = err.Error()
		audit.Log(err, "change-password", lres.SError, "user", "", "email", "")

		return &lres, nil
	}

	pass := r.FormValue("password")
	newPass := r.FormValue("new_password")
	confirmPass := r.FormValue("confirm_password")

	if len(pass) == 0 {
		lres.BError = true
		lres.SError = "Old password cannot be empty"
		err = fmt.Errorf(lres.SError)
		audit.Log(err, "change-password", lres.SError, "user", usr.Username, "email", usr.Email)

		return &lres, nil
	}

	if len(newPass) == 0 || newPass != confirmPass {
		lres.BError = true
		lres.SError = "Password is empty or is different from it's confirmation."
		err = fmt.Errorf(lres.SError)
		audit.Log(err, "change-password", lres.SError, "user", usr.Username, "email", usr.Email)

		return &lres, nil
	}

	if pass == newPass {
		lres.BError = true
		lres.SError = "The new password must be different from the current one."
		err = fmt.Errorf(lres.SError)
		audit.Log(err, "change-password", lres.SError, "user", usr.Username, "email", usr.Email)

		return &lres, nil
	}

	ip := getClientIP(r)

	success, err := ValidateUserPassword(usr.Username, pass, ip)

	if success != ValidationOK && success != ValidationTemporaryPassword {
		lres.BError = true
		lres.SError = "Old password is not valid."
		err = fmt.Errorf(lres.SError)
		audit.Log(err, "change-password", lres.SError, "user", usr.Username, "email", usr.Email)

		return &lres, nil
	}

	usr.Password = newPass

	err = usr.Save()

	if err != nil {
		lres.BError = true
		lres.SError = err.Error()
		audit.Log(err, "change-password", lres.SError, "user", usr.Username, "email", usr.Email)

		return &lres, err
	}

	if sessionData.User.TempPassword {
		sessionData.User.TempPassword = false

		err = refreshSessionData(w, r, *sessionData)
		if err != nil {
			lres.BError = true
			lres.SError = err.Error()
			audit.Log(err, "change-password", lres.SError, "user", usr.Username, "email", usr.Email)

			return &lres, err
		}
	}

	lres.BError = false
	lres.SError = "User password changed"
	audit.Log(nil, "change-password", lres.SError, "user", usr.Username, "email", usr.Email)

	return &lres, nil
}

func (HomeController) GetExchangeRates(w http.ResponseWriter, r *http.Request, res *ResponseHelper) (*models.ExchangeRatesResponseModel, error) {
	var lres models.ExchangeRatesResponseModel
	var date string

	if val, ok := r.Form["date"]; ok {
		date = val[0]
	}

	if !utils.IsISODate(date) {
		date = ""
	}

	queryAux := "select max(exchange_date) from exchange_rate"

	if len(date) > 0 {
		queryAux += " where exchange_date <= DATE ? "
	}

	query := dbUtils.PQuery(`
		select c.currency, r.exchange_date, r.rate
		  from exchange_rate r
		  join currency c on (r.currency_id = c.currency_id)
		 where exchange_date = (` + queryAux + `)
		 order by c.currency, r.exchange_date
	`)

	var err error
	var rows *sql.Rows

	if len(date) > 0 {
		rows, err = db.Query(query, date)
	} else {
		rows, err = db.Query(query)
	}
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

	rows.Close()

	return &lres, nil
}
