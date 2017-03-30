package main

import (
	"strings"
)

type SystemParams struct {
	Group  string
	Params map[string]string
}

func (p *SystemParams) LoadByGroup(group string) error {
	p.Group = strings.ToLower(group)
	p.Params = make(map[string]string)

	query := `
		SELECT param, val FROM wmeter.system_params WHERE param_group = $1
	`

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
