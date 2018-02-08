package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/geo-stanciu/go-utils/utils"

	_ "github.com/mattn/go-oci8"
)

var (
	db     *sql.DB
	config = Configuration{}
	dbutl  *utils.DbUtils
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

	loc, err := time.LoadLocation("Europe/Bucharest")
	if err != nil {
		panic(err)
	}

	test := Test{}
	pq := dbutl.PQuery(`
		select current_timestamp "date", '12.2.1.0' version from dual
	`)

	err = dbutl.RunQuery(pq, &test)
	if err != nil {
		panic(err)
	}

	fmt.Println("Date:", test.Date)
	fmt.Println("Date - local:", test.Date.In(loc))
	//fmt.Println(test.Version)

	err = dbutl.ForEachRow(pq, func(row *sql.Rows, sc *utils.SQLScan) error {
		test2 := Test{}
		err = sc.Scan(dbutl, row, &test2)
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

	query := `
	    select CASE WHEN EXISTS (
			select 1 from user_tables where table_name = 'TEST1'
		) THEN 1 ELSE 0 END
		FROM dual
	`

	found := false
	err = db.QueryRow(query).Scan(&found)
	if err != nil {
		panic(err)
	}

	if !found {
		query := `
	    create table test1 (
			dt date,
			dtz timestamp,
			d date,
			d_null date
		)
	`

		_, err = db.Exec(query)
		if err != nil {
			panic(err)
		}
	}

	query = `
	    select CASE WHEN EXISTS (
			select 1 from test1
		) THEN 1 ELSE 0 END
		FROM dual
	`

	err = db.QueryRow(query).Scan(&found)
	if err != nil {
		panic(err)
	}

	if !found {
		now := time.Now().UTC()

		pq = dbutl.PQuery(`
			insert into test1 (
				dt,
				dtz,
				d
			)
			values (?, ?, ?)
		`, now, now, now)

		_, err = dbutl.Exec(pq)
		if err != nil {
			panic(err)
		}
	}

	pq = dbutl.PQuery(`select dt, dtz, d, d_null from test1 order by 1`)

	err = dbutl.ForEachRow(pq, func(row *sql.Rows, sc *utils.SQLScan) error {
		test1 := Test1{}
		err = sc.Scan(dbutl, row, &test1)
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
