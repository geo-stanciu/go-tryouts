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
		"geo",
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

	err = db.Ping()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("opened...")

	now, err := currentTime(db)

	if err != nil {
		log.Fatal(err)
	}

	count, err := getStatCount(db)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Date: %s - Stat count: %d\n", now, count)
}

func getStatCount(db *sql.DB) (int64, error) {
	var count int64

	query := "select count(*) from pg_stat_activity where usename = $1"

	err := db.QueryRow(query, "geo").Scan(&count)

	switch {
	case err == sql.ErrNoRows:
		return 0, err
	case err != nil:
		return 0, err
	default:
		return count, nil
	}
}

func currentTime(db *sql.DB) (string, error) {
	var currentTime string

	rows, err := db.Query("select current_timestamp")

	if err != nil {
		return "", err
	}

	defer rows.Close()

	for rows.Next() {
		var date time.Time
		err = rows.Scan(&date)

		if err != nil {
			return "", err
		}

		// the format is interpreted by passing date value 2006-01-02 15:04:05.000
		// in your desired format
		currentTime = date.Format("2006-01-02 15:04:05.000")
	}

	if err := rows.Err(); err != nil {
		return "", err
	}

	return currentTime, nil
}
