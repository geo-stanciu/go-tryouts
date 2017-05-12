package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

func main() {
	db, err := sql.Open("mysql", "geo:geo@tcp(127.0.0.1:3306)/devel?parseTime=true&loc=Europe%2FBucharest&sql_mode=TRADITIONAL")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	rows, err := db.Query("select current_timestamp")
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
