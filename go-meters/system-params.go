package main

import (
	"database/sql"
	"strconv"
	"strings"
)

func getSystemParamAsInt(key string, defaultVal string) int {
	sval := getSystemParam(key, defaultVal)

	val, err := strconv.Atoi(sval)

	if err != nil {
		Log(true, err, "system-params", "Conversion error.", "param", key)
		return 0
	}

	return val
}

func getSystemParam(key string, defaultVal string) string {
	var rez string

	query := `
		SELECT val FROM wmeter.system_params WHERE param = $1
	`

	err := db.QueryRow(query, strings.ToLower(key)).Scan(&rez)

	switch {
	case err == sql.ErrNoRows:
		Log(true, nil, "system-params", "param not found", "param", key)
		return defaultVal
	case err != nil:
		Log(true, nil, "system-params", "error getting param value", "param", key)
		return defaultVal
	}

	return rez
}
