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
	dbUtils = utils.DbUtils{}
)

type Test struct {
	Date    time.Time
	Version string
}

type Test1 struct {
	Dt  time.Time
	Dtz time.Time
	D   time.Time
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

	test := Test{}
	query := dbUtils.PQuery("select current_timestamp date, @@version version")
	err = utils.RunQuery(db, query, &test)
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

	fmt.Println(test.Date)
	fmt.Println(test.Date.In(loc))
	fmt.Println(test.Version)

	sc := utils.SQLScanHelper{}

	err = utils.ForEachRow(db, query, func(row *sql.Rows) {
		test2 := Test{}
		err = sc.Scan(row, &test2)
		if err != nil {
			panic(err)
		}

		fmt.Println(test2.Date)
		fmt.Println(test2.Date.In(loc))
		fmt.Println(test2.Version)
	})

	if err != nil {
		panic(err)
	}

	query = `
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

	/*query = dbUtils.PQuery(`
		insert into test1 (
			dt,
			dtz,
			d
		)
		values (?, ?, ?)
	`)

	now = time.Now().UTC()
	_, err = db.Exec(query, now, now, now)

	if err != nil {
		panic(err)
	}*/

	query = dbUtils.PQuery(`select dt, dtz, d from test1 order by 1`)

	sc.Clear()
	err = utils.ForEachRow(db, query, func(row *sql.Rows) {
		test1 := Test1{}
		err = sc.Scan(row, &test1)
		if err != nil {
			panic(err)
		}

		fmt.Println(test1.Dt)
		fmt.Println(test1.Dt.In(loc))
		fmt.Println(test1.Dtz)
		fmt.Println(test1.D)
	})

	if err != nil {
		panic(err)
	}
}
