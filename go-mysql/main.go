package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/geo-stanciu/go-utils/utils"

	_ "github.com/go-sql-driver/mysql"
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

	loc, err := time.LoadLocation("Europe/Bucharest")
	if err != nil {
		panic(err)
	}

	test := Test{}
	query := dbUtils.PQuery("select current_timestamp date, version() as version")
	err = dbUtils.RunQuery(query, &test)
	if err != nil {
		panic(err)
	}

	fmt.Println("Date: ", test.Date)
	fmt.Println("Date - local: ", test.Date.In(loc))
	//fmt.Println(test.Version)

	sc := utils.SQLScanHelper{}
	err = dbUtils.ForEachRow(query, func(row *sql.Rows) {
		test2 := Test{}
		err = sc.Scan(&dbUtils, row, &test2)
		if err != nil {
			panic(err)
		}

		fmt.Println("Date: ", test2.Date)
		fmt.Println("Date - local:", test2.Date.In(loc))
	})

	if err != nil {
		panic(err)
	}

	query = `
		create table if not exists test1 (
			dt datetime(3),
			dtz timestamp,
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

	now := time.Now().UTC()
	_, err = db.Exec(query, now, now, now)

	if err != nil {
		panic(err)
	}*/

	query = dbUtils.PQuery(`select dt, dtz, d from test1 order by 1`)

	sc.Clear()
	err = dbUtils.ForEachRow(query, func(row *sql.Rows) {
		test1 := Test1{}
		err = sc.Scan(&dbUtils, row, &test1)
		if err != nil {
			panic(err)
		}

		fmt.Println("Dt:", test1.Dt)
		fmt.Println("Dt - local:", test1.Dt.In(loc))
		fmt.Println("Dtz:", test1.Dtz)
		fmt.Println("D:", test1.D)
	})

	if err != nil {
		panic(err)
	}
}
