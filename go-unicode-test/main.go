package main

import (
	"database/sql"
	"fmt"

	"github.com/geo-stanciu/go-utils/utils"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-oci8"
)

var (
	db      *sql.DB
	config  = Configuration{}
	dbUtils *utils.DbUtils
)

func init() {
	// init databaseutils
	dbUtils = new(utils.DbUtils)
}

func main() {
	var err error

	cfgFile := "./conf.json"
	err = config.ReadFromFile(cfgFile)
	if err != nil {
		panic(err)
	}

	err = dbUtils.Connect2Database(&db, config.DbType, config.DbURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	pq := dbUtils.PQuery(`
		DELETE FROM test_unicode
	`)

	_, err = tx.Exec(pq.Query)
	if err != nil {
		panic(err)
	}

	pq = dbUtils.PQuery(`
		INSERT INTO test_unicode (c1) VALUES (?)
	`, "Hello, 世界")

	_, err = tx.Exec(pq.Query, pq.Args...)
	if err != nil {
		panic(err)
	}

	pq = dbUtils.PQuery(`SELECT c1 FROM test_unicode`)

	err = dbUtils.ForEachRowTx(tx, pq, func(row *sql.Rows) {
		var c1 string
		err = row.Scan(&c1)
		if err != nil {
			return
		}

		fmt.Println("c1: ", c1)
	})

	tx.Commit()
}
