package main

import (
	"database/sql"
	"fmt"

	"github.com/geo-stanciu/go-utils/utils"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	//_ "github.com/mattn/go-oci8"
)

var (
	db     *sql.DB
	config = Configuration{}
	dbutl  *utils.DbUtils
)

func init() {
	// init databaseutils
	dbutl = new(utils.DbUtils)
}

func main() {
	var err error

	cfgFile := "./conf.json"
	err = config.ReadFromFile(cfgFile)
	if err != nil {
		panic(err)
	}

	err = dbutl.Connect2Database(&db, config.DbType, config.DbURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	pq := dbutl.PQuery(`
		DELETE FROM test_unicode
	`)

	_, err = tx.Exec(pq.Query)
	if err != nil {
		panic(err)
	}

	pq = dbutl.PQuery(`
		INSERT INTO test_unicode (c1) VALUES (?)
	`, "Hello, 世界")

	_, err = tx.Exec(pq.Query, pq.Args...)
	if err != nil {
		panic(err)
	}

	pq = dbutl.PQuery(`SELECT c1 FROM test_unicode`)

	err = dbutl.ForEachRowTx(tx, pq, func(row *sql.Rows, sc *utils.SQLScan) error {
		var c1 string
		err = row.Scan(&c1)
		if err != nil {
			return err
		}

		fmt.Println("c1: ", c1)

		return nil
	})

	tx.Commit()
}
