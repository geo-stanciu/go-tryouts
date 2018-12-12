package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/geo-stanciu/go-utils/utils"

	_ "github.com/denisenkom/go-mssqldb"
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

	t1 := test{}
	pq := dbutl.PQuery("select current_timestamp date, @@version version")
	err = dbutl.RunQuery(pq, &t1)
	if err != nil {
		panic(err)
	}

	fmt.Println("Date:", t1.Date)
	fmt.Println("Date - local:", t1.Date.In(loc))

	err = dbutl.ForEachRow(pq, func(row *sql.Rows, sc *utils.SQLScan) error {
		t2 := test{}
		err = sc.Scan(dbutl, row, &t2)
		if err != nil {
			return err
		}

		fmt.Println("Date:", t2.Date)
		fmt.Println("Date - local:", t2.Date.In(loc))

		return nil
	})

	if err != nil {
		panic(err)
	}

	query := `
		if not exists (select * from sysobjects where xtype = 'U' and name = 'test1')
		create table test1 (
			dt datetime,
			dtz datetime,
			d date,
			d_null datetime
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

	found := 0
	err = db.QueryRow(query).Scan(&found)
	if err != nil {
		panic(err)
	}

	if found == 0 {
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

		fmt.Println("Dt: ", t1.Dt)
		fmt.Println("Dt - local: ", t1.Dt.In(loc))
		fmt.Println("Dtz: ", t1.Dtz)
		fmt.Println("D: ", t1.D)

		if t1.DNull.Valid {
			fmt.Println("D NUll:", t1.DNull.Time)
		} else {
			fmt.Println("D NUll: null")
		}

		return nil
	})

	if err != nil {
		panic(err)
	}
}
