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
	dbUtils = utils.DbUtils{}
)

func main() {
	var err error

	//cfgFile := "./conf.json"
	cfgFile := "./conf_SQLSRV.json"
	err = config.ReadFromFile(cfgFile)
	if err != nil {
		panic(err)
	}

	err = dbUtils.Connect2Database(&db, config.DbType, config.DbURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	query := dbUtils.PQuery(`
		INSERT INTO test_unicode (c1) VALUES (?)
	`)

	_, err = db.Exec(query, "Hello, 世界")
	if err != nil {
		panic(err)
	}

	query = dbUtils.PQuery(`SELECT c1 FROM test_unicode`)

	err = dbUtils.ForEachRow(query, func(row *sql.Rows) {
		var c1 string
		err = row.Scan(&c1)
		if err != nil {
			return
		}

		fmt.Println("c1: ", c1)
	})
}
