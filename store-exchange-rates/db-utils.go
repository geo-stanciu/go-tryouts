package main

import (
	"fmt"
	"strings"
)

func prepareQuery(query string) string {
	q := query
	dbType := strings.ToLower(config.DbType)

	i := 1
	prefix := ""

	if dbType == "postgres" {
		prefix = "$"
	}

	if dbType != "mysql" {
		for {
			idx := strings.Index(q, "?")

			if idx < 0 {
				break
			}

			prm := fmt.Sprintf("%s%d", prefix, i)
			i++

			q = strings.Replace(q, "?", prm, 1)
		}
	}

	return q
}
