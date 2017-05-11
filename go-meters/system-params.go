package main

import (
	"strings"
)

const (
	PasswordRules             string = "password-rules"
	ChangeInterval            string = "change-interval"
	PasswordFailInterval      string = "password-fail-interval"
	MaxAllowedFailedAtmpts    string = "max-allowed-failed-atmpts"
	NotRepeatLastXPasswords   string = "not-repeat-last-x-passwords"
	MinCharacters             string = "min-characters"
	MinLetters                string = "min-letters"
	MinCapitals               string = "min-capitals"
	MinDigits                 string = "min-digits"
	MinNonAlphaNumerics       string = "min-non-alpha-numerics"
	AllowRepetitiveCharacters string = "allow-repetitive-characters"
	CanContainUsername        string = "can-contain-username"
)

type SystemParams struct {
	Group  string
	Params map[string]string
}

func (p *SystemParams) LoadByGroup(group string) error {
	p.Group = strings.ToLower(group)
	p.Params = make(map[string]string)

	query := dbUtils.PQuery(`
		SELECT param, val
		  FROM system_params
		 WHERE param_group = ?
		 ORDER BY param
	`)

	rows, err := db.Query(query, p.Group)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var key string
		var val string

		err = rows.Scan(&key, &val)
		if err != nil {
			return err
		}

		p.Params[key] = val
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

func (p SystemParams) GetString(key string) string {
	return p.Params[key]
}

func (p SystemParams) GetInt(key string) int {
	return string2int(p.Params[key])
}
