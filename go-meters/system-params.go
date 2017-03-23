package main

import (
	"database/sql"
	"strings"
)

func getAllParamsByGroup(group string) (map[string]string, error) {
	rez := make(map[string]string)

	query := `
		SELECT key, val FROM wmeter.system_params WHERE param_group = $1
	`

	rows, err := db.Query(query)

	if err != nil {
		return rez, err
	}

	defer rows.Close()

	for rows.Next() {
		var key string
		var val string

		err = rows.Scan(&key, &val)

		if err != nil {
			return rez, err
		}

		rez[key] = val
	}

	if err := rows.Err(); err != nil {
		return rez, err
	}

	return rez, nil
}

func getSystemParam(group string, key string, defaultVal string) string {
	var rez string

	query := `
		SELECT val FROM wmeter.system_params WHERE param_group = $1, param = $2
	`

	err := db.QueryRow(
		query,
		strings.ToLower(group),
		strings.ToLower(key),
	).Scan(&rez)

	switch {
	case err == sql.ErrNoRows:
		Log(true, nil, "system-params", "param not found", "group", group, "param", key)
		return defaultVal
	case err != nil:
		Log(true, nil, "system-params", "error getting param value", "group", group, "param", key)
		return defaultVal
	}

	return rez
}
