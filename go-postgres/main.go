package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/geo-stanciu/go-utils/utils"

	_ "github.com/lib/pq"
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
	query := dbUtils.PQuery("select current_timestamp date, version() as version")
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

	//time.Sleep(1 * time.Second)

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
}
