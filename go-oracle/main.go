package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/geo-stanciu/go-utils/utils"

	_ "github.com/mattn/go-oci8"
)

var (
	db      *sql.DB
	config  = Configuration{}
	dbUtils *utils.DbUtils
)

type Test struct {
	Date    time.Time `sql:"date"`
	Version string    `sql:"version"`
}

type Test1 struct {
	Dt  time.Time `sql:"dt"`
	Dtz time.Time `sql:"dtz"`
	D   time.Time `sql:"d"`
}

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

	loc, err := time.LoadLocation("Europe/Bucharest")
	if err != nil {
		panic(err)
	}

	test := Test{}
	pq := dbUtils.PQuery(`
		select current_timestamp "date", '12.2.1.0' version from dual
	`)

	err = dbUtils.RunQuery(pq, &test)
	if err != nil {
		panic(err)
	}

	fmt.Println("Date:", test.Date)
	fmt.Println("Date - local:", test.Date.In(loc))
	fmt.Println(test.Version)

	err = dbUtils.ForEachRow(pq, func(row *sql.Rows, sc *utils.SQLScanHelper) error {
		test2 := Test{}
		err = sc.Scan(dbUtils, row, &test2)
		if err != nil {
			return err
		}

		fmt.Println("Date:", test2.Date)
		fmt.Println("Date - local:", test2.Date.In(loc))
		//fmt.Println(test2.Version)

		return nil
	})

	if err != nil {
		panic(err)
	}

	/*query := `
			create table test1 (
				dt date,
				dtz timestamp,
				d date
			)
		`

	_, err = db.Exec(query)

	if err != nil {
		panic(err)
	}*/

	/*
		now = time.Now().UTC()

		pq = dbUtils.PQuery(`
			insert into test1 (
				dt,
				dtz,
				d
			)
			values (?, ?, ?)
		`, now, now, now)

		_, err = dbUtils.Exec(pq)
		if err != nil {
			panic(err)
		}*/

	/*pq = dbUtils.PQuery(`select dt, dtz, d from test1 order by 1`)

	sc.Clear()
	err = dbUtils.ForEachRow(pq, func(row *sql.Rows, sc *utils.SQLScanHelper) error {
		test1 := Test1{}
		err = sc.Scan(dbUtils, row, &test1)
		if err != nil {
			return err
		}

		fmt.Println(test1.Dt)
		fmt.Println(test1.Dt.In(loc))
		fmt.Println(test1.Dtz)
		fmt.Println(test1.D)

		return nil
	})

	if err != nil {
		panic(err)
	}*/
}
