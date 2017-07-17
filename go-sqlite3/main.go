package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	//"os"
)

func main() {
	//os.Remove("./foo.db")

	//db, err := sql.Open("sqlite3", "./foo.db")
	db, err := sql.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()

	sqlStmt := `
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

	rows.Close()
}
