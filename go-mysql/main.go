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
	dbUtils *utils.DbUtils
)

type Test struct {
	Date    time.Time `sql:"date"`
	Version string    `sql:"version"`
}

type Test1 struct {
	Dt    time.Time      `sql:"dt"`
	Dtz   time.Time      `sql:"dtz"`
	D     time.Time      `sql:"d"`
	DNull utils.NullTime `sql:"d_null"`
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
	pq := dbUtils.PQuery("select current_timestamp date, version() as version")
	err = dbUtils.RunQuery(pq, &test)
	if err != nil {
		panic(err)
	}

	fmt.Println("Date: ", test.Date)
	fmt.Println("Date - local: ", test.Date.In(loc))
	//fmt.Println(test.Version)

	err = dbUtils.ForEachRow(pq, func(row *sql.Rows, sc *utils.SQLScan) error {
		test2 := Test{}
		err = sc.Scan(dbUtils, row, &test2)
		if err != nil {
			return err
		}

		fmt.Println("Date: ", test2.Date)
		fmt.Println("Date - local:", test2.Date.In(loc))
		fmt.Println("Version: ", test2.Version)

		return nil
	})

	if err != nil {
		panic(err)
	}

	query := `
		create table if not exists test1 (
			dt datetime(3),
			dtz timestamp,
			d date,
			d_null datetime(3)
		)
	`

	_, err = db.Exec(query)
	if err != nil {
		panic(err)
	}

	query = `
	    select CASE WHEN EXISTS (
			select 1 from test1
		) THEN 1 ELSE 0 END
	`

	found := false
	err = db.QueryRow(query).Scan(&found)
	if err != nil {
		panic(err)
	}

	if !found {
		now := time.Now().UTC()

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
		}
	}

	pq = dbUtils.PQuery(`select dt, dtz, d, d_null from test1 order by 1`)

	err = dbUtils.ForEachRow(pq, func(row *sql.Rows, sc *utils.SQLScan) error {
		test1 := Test1{}
		err = sc.Scan(dbUtils, row, &test1)
		if err != nil {
			return err
		}

		fmt.Println("Dt:", test1.Dt)
		fmt.Println("Dt - local:", test1.Dt.In(loc))
		fmt.Println("Dtz:", test1.Dtz)
		fmt.Println("D:", test1.D)

		if test1.DNull.Valid {
			fmt.Println("D NUll:", test1.DNull.Time)
		} else {
			fmt.Println("D NUll: null")
		}

		return nil
	})

	if err != nil {
		panic(err)
	}
}
