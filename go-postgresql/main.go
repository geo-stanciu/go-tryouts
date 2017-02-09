package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&application_name=%s",
		"geo",
		"p",
		"localhost",
		"5432",
		"devel",
		"disable",
		"go-test-postgresql")

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	fmt.Println("opened...")

	rows, err := db.Query("select current_timestamp")

	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	for rows.Next() {
		var date time.Time
		err = rows.Scan(&date)

		if err != nil {
			log.Fatal(err)
		}

		// the format is interpreted by passing date value 2006-01-02 15:04:05.000
		// in your desired format
		fmt.Printf("Date: %v\n", date.Format("2006-01-02 15:04:05.000"))
	}
}
