package main

import (
	"database/sql"
	"strings"

	"github.com/geo-stanciu/go-utils/utils"
)

const (
	// PasswordRules - password rules group
	PasswordRules string = "password-rules"
	// ChangeInterval - password change interval
	ChangeInterval string = "change-interval"
	// PasswordFailInterval - password fail interval
	PasswordFailInterval string = "password-fail-interval"
	// MaxAllowedFailedAtmpts - max allowed failed atmpts
	MaxAllowedFailedAtmpts string = "max-allowed-failed-atmpts"
	// NotRepeatLastXPasswords - not repeat last x passwords
	NotRepeatLastXPasswords string = "not-repeat-last-x-passwords"
	// MinCharacters - min characters for passwords
	MinCharacters string = "min-characters"
	// MinLetters - min letters for passwords
	MinLetters string = "min-letters"
	// MinCapitals - min capitals for passwords
	MinCapitals string = "min-capitals"
	// MinDigits - min digits for passwords
	MinDigits string = "min-digits"
	// MinNonAlphaNumerics - min non alphaNumerics for passwords
	MinNonAlphaNumerics string = "min-non-alpha-numerics"
	// AllowRepetitiveCharacters - allow repetitive characters for passwords
	AllowRepetitiveCharacters string = "allow-repetitive-characters"
	// CanContainUsername - can contain username for passwords
	CanContainUsername string = "can-contain-username"
)

// SystemParams - structure helper for system settings
type SystemParams struct {
	Group  string
	Params map[string]string
}

// LoadByGroup - load settings by group
func (p *SystemParams) LoadByGroup(group string) error {
	p.Group = strings.ToLower(group)
	p.Params = make(map[string]string)

	query := dbUtils.PQuery(`
		SELECT param, val
		  FROM system_params
		 WHERE param_group = ?
		 ORDER BY param
	`)

	var err error
	err = dbUtils.ForEachRow(query, func(row *sql.Rows) {
		var key string
		var val string

		err = row.Scan(&key, &val)
		if err != nil {
			return
		}

		p.Params[key] = val
	}, p.Group)

	if err != nil {
		return err
	}

	return nil
}

// GetString - Get setting as a string
func (p SystemParams) GetString(key string) string {
	return p.Params[key]
}

// GetInt - Get setting as an int
func (p SystemParams) GetInt(key string) int {
	return utils.String2int(p.Params[key])
}
