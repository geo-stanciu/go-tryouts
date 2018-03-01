package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/geo-stanciu/go-utils/utils"

	_ "github.com/mattn/go-sqlite3"
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
	pq := dbutl.PQuery("select current_timestamp date, sqlite_version() as version")
	err = dbutl.RunQuery(pq, &t1)
	if err != nil {
		panic(err)
	}

	fmt.Println("Date: ", t1.Date)
	fmt.Println("Date - local: ", t1.Date.In(loc))
	//fmt.Println(t1.Version)

	/*sqlStmt := `
		create table foo (id integer not null primary key, name text);
		delete from foo;
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		return
	}

	stmt, err := tx.Prepare("insert into foo(id, name) values(?, ?)")
	if err != nil {
		log.Println(err)
		return
	}
	defer stmt.Close()

	for i := 0; i < 100; i++ {
		_, err = stmt.Exec(i, fmt.Sprintf("%03d", i))
		if err != nil {
			log.Println(err)
			return
		}
	}

	tx.Commit()

	rows, err := db.Query("select id, name from foo")
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Println(id, name)
	}

	err = rows.Err()
	if err != nil {
		log.Println(err)
		return
	}

	rows.Close()

	stmt, err = db.Prepare("select name from foo where id = ?")
	if err != nil {
		log.Println(err)
		return
	}
	defer stmt.Close()

	var name string
	err = stmt.QueryRow("3").Scan(&name)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(name)

	_, err = db.Exec("delete from foo")
	if err != nil {
		log.Println(err)
		return
	}

	_, err = db.Exec("insert into foo(id, name) values(1, 'foo'), (2, 'bar'), (3, 'baz')")
	if err != nil {
		log.Println(err)
		return
	}

	rows, err = db.Query("select id, name from foo")
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Println(id, name)
	}

	err = rows.Err()
	if err != nil {
		log.Println(err)
		return
	}

	rows.Close()

	rows, err = db.Query("select strftime('%Y-%m-%d %H:%M:%S', current_timestamp, 'localtime')")
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var dt string
		err = rows.Scan(&dt)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Println(dt)
	}

	err = rows.Err()
	if err != nil {
		log.Println(err)
		return
	}

	rows.Close()*/
}
