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
	config = configuration{}
	dbutl  *utils.DbUtils
)

type test struct {
	Date    time.Time `sql:"date"`
	Version string    `sql:"version"`
}

type test1 struct {
	Dt    time.Time      `sql:"dt"`
	Dtz   time.Time      `sql:"dtz"`
	D     time.Time      `sql:"d"`
	DNull utils.NullTime `sql:"d_null"`
}

type tablename struct {
	Tname string `sql:"table_name"`
}

func init() {
	// init databaseutils
	dbutl = new(utils.DbUtils)
}

func main() {
	var err error

	cfgFile := "./conf.json"
	err = config.readFromFile(cfgFile)
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

	t1 := test{}
	pq := dbutl.PQuery(`
		select current_timestamp "date", '12.2.1.0' version from dual
	`)

	err = dbutl.RunQuery(pq, &t1)
	if err != nil {
		panic(err)
	}

	fmt.Println("Date:", t1.Date)
	fmt.Println("Date - local:", t1.Date.In(loc))
	//fmt.Println(t1.Version)

	err = dbutl.ForEachRow(pq, func(row *sql.Rows, sc *utils.SQLScan) error {
		t2 := test{}
		err = sc.Scan(dbutl, row, &t2)
		if err != nil {
			return err
		}

		fmt.Println("Date:", t2.Date)
		fmt.Println("Date - local:", t2.Date.In(loc))
		//fmt.Println(t2.Version)

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
		t1 := test1{}
		err = sc.Scan(dbutl, row, &t1)
		if err != nil {
			return err
		}

		fmt.Println("Dt:", t1.Dt)
		fmt.Println("Dt - local:", t1.Dt.In(loc))
		fmt.Println("Dtz:", t1.Dtz)
		fmt.Println("D:", t1.D)

		if t1.DNull.Valid {
			fmt.Println("D NUll:", t1.DNull.Time)
		} else {
			fmt.Println("D NUll: null")
		}

		return nil
	})

	pq = dbutl.PQuery(`
		select table_name
		  from all_tables
		 order by 1
		 limit ? offset ?
	`, 2, 1)

	err = dbutl.ForEachRow(pq, func(row *sql.Rows, sc *utils.SQLScan) error {
		t1 := tablename{}
		err = sc.Scan(dbutl, row, &t1)
		if err != nil {
			return err
		}

		fmt.Println("table:", t1.Tname)

		return nil
	})

	if err != nil {
		panic(err)
	}
}
