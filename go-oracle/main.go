package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-oci8"
)

func main() {
	db, err := sql.Open("oci8", "geo/geo@db1")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	rows, err := db.Query("select current_timestamp from dual")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var dt time.Time
		err = rows.Scan(&dt)
		if err != nil {
			panic(err)
		}

		fmt.Println(dt)
		fmt.Println(dt.UTC())
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}
}
