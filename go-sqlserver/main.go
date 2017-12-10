package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

func main() {
	db, err := sql.Open("mssql", `server=localhost;database=devel;user id=geo;password=geo;port=1433;app name=Test1`)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	rows, err := db.Query("select GETUTCDATE()")
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

		now := time.Now()

		fmt.Println(now)
		fmt.Println(now.UTC())

		loc, err := time.LoadLocation("Europe/Bucharest")
		if err != nil {
			panic(err)
		}

		fmt.Println(dt)
		fmt.Println(dt.In(loc))
	}

	err = rows.Err()
	if err != nil {
		panic(err)
	}

	rows.Close()
}
