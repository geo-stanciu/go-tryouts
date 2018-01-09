package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/geo-stanciu/go-utils/utils"

	_ "github.com/denisenkom/go-mssqldb"
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
	pq := dbUtils.PQuery("select current_timestamp date, @@version version")
	err = dbUtils.RunQuery(pq, &test)
	if err != nil {
		panic(err)
	}

	fmt.Println("Date:", test.Date)
	fmt.Println("Date - local:", test.Date.In(loc))

	sc := utils.SQLScanHelper{}
	err = dbUtils.ForEachRow(pq, func(row *sql.Rows) {
		test2 := Test{}
		err = sc.Scan(dbUtils, row, &test2)
		if err != nil {
			panic(err)
		}

		fmt.Println("Date:", test2.Date)
		fmt.Println("Date - local:", test2.Date.In(loc))
	})

	if err != nil {
		panic(err)
	}

	query := `
		if not exists (select * from sysobjects where xtype = 'U' and name = 'test1')
		create table test1 (
			dt datetime,
			dtz datetime,
			d date
		)
	`

	_, err = db.Exec(query)

	if err != nil {
		panic(err)
	}

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

	pq = dbUtils.PQuery(`select dt, dtz, d from test1 order by 1`)

	sc.Clear()
	err = dbUtils.ForEachRow(pq, func(row *sql.Rows) {
		test1 := Test1{}
		err = sc.Scan(dbUtils, row, &test1)
		if err != nil {
			panic(err)
		}

		fmt.Println("Dt: ", test1.Dt)
		fmt.Println("Dt - local: ", test1.Dt.In(loc))
		fmt.Println("Dtz: ", test1.Dtz)
		fmt.Println("D: ", test1.D)
	})

	if err != nil {
		panic(err)
	}
}
