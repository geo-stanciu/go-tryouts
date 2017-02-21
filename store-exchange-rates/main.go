package main

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

type Rate struct {
	Currency   string `xml:"currency,attr"`
	Multiplier string `xml:"multiplier,attr"`
	Rate       string `xml:",chardata"`
}

type Cube struct {
	Date string `xml:"date,attr"`
	Rate []Rate
}

type Header struct {
	Publisher      string `xml:"Publisher"`
	PublishingDate string `xml:"PublishingDate"`
	MessageType    string `xml:"MessageType"`
}

type Body struct {
	Subject      string `xml:"Subject"`
	OrigCurrency string `xml:"OrigCurrency"`
	Cube         []Cube
}

type Query struct {
	XMLName xml.Name `xml:"DataSet"`
	Header  Header   `xml:"Header"`
	Body    Body     `xml:"Body"`
}

func main() {
	var xmlBytes []byte
	var err error

	db, err := connect2Database()

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	err = db.Ping()

	if err != nil {
		log.Fatal(err)
	}

	err = createTablesIfNotExist(db)

	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) >= 2 {
		xmlBytes, err = readBytesFromFile(os.Args[1])
	} else {
		xmlBytes, err = readBytesFromURL("http://bnro.ro/nbrfxrates.xml")
	}

	if err != nil {
		log.Fatal(err)
	}

	err = dealWithXML(db, xmlBytes)

	if err != nil {
		log.Fatal(err)
	}
}

func readBytesFromURL(url string) ([]byte, error) {
	response, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	buf, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	return buf, nil
}

func readBytesFromFile(filename string) ([]byte, error) {
	f, err := os.Open(filename)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	buf, err := ioutil.ReadAll(f)

	if err != nil {
		return nil, err
	}

	return buf, nil
}

func connect2Database() (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&application_name=%s",
		"geo",
		"geo",
		"localhost",
		"5432",
		"devel",
		"disable",
		"go-test-postgresql")

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, err
	}

	return db, nil
}

func createTablesIfNotExist(db *sql.DB) error {
	t1 := `
		create table if not exists currency (
			currency_id serial primary key,
			currency    varchar(8) not null,
			constraint currency_uk unique (currency)
		)
	`

	t2 := `
		create table if not exists exchange_rate (
			exchange_rate_id serial primary key,
			currency_id      int            not null,
			exchange_date    date           not null,       
			rate             numeric(18, 6) not null,
			constraint exchange_rate_currency_fk foreign key (currency_id)
			    references currency (currency_id),
			constraint exchange_rate_uk unique (currency_id, exchange_date)
		)
	`

	_, err := db.Exec(t1)

	if err != nil {
		return err
	}

	_, err = db.Exec(t2)

	if err != nil {
		return err
	}

	var count int32

	tx, err := db.Begin()

	if err != nil {
		return err
	}

	defer tx.Rollback()

	query := "select count(*) from currency"

	err = tx.QueryRow(query).Scan(&count)

	switch {
	case err == sql.ErrNoRows:
		return err
	case err != nil:
		return err
	}

	if count == 0 {
		query = "insert into currency (currency) values ($1)"

		_, err = tx.Exec(query, "RON")

		if err != nil {
			return err
		}

		_, err = tx.Exec(query, "EUR")

		if err != nil {
			return err
		}

		_, err = tx.Exec(query, "USD")

		if err != nil {
			return err
		}

		_, err = tx.Exec(query, "CHF")

		if err != nil {
			return err
		}
	}

	tx.Commit()

	return nil
}

func dealWithXML(db *sql.DB, xmlBytes []byte) error {
	var q Query

	err := xml.Unmarshal(xmlBytes, &q)

	if err != nil {
		return err
	}

	tx, err := db.Begin()

	if err != nil {
		return err
	}

	defer tx.Rollback()

	for _, cube := range q.Body.Cube {
		fmt.Printf("Importing exchange rates for %s\n", cube.Date)

		for _, rate := range cube.Rate {
			multiplier := 1.0
			exchRate := 1.0

			if len(rate.Multiplier) > 0 {
				multiplier, err = strconv.ParseFloat(rate.Multiplier, 64)

				if err != nil {
					return err
				}
			}

			exchRate, err = strconv.ParseFloat(rate.Rate, 64)

			if err != nil {
				return err
			}

			err = storeRate(tx, cube.Date, rate.Currency, multiplier, exchRate)

			if err != nil {
				return err
			}
		}
	}

	tx.Commit()

	return nil
}

func storeRate(tx *sql.Tx, date string, currency string, multiplier float64, exchRate float64) error {
	var currencyID int32
	var count int32

	query := "select currency_id from currency where currency = $1"

	err := tx.QueryRow(query, currency).Scan(&currencyID)

	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		return err
	}

	query = `
		select count(*) 
		  from exchange_rate 
		where currency_id = $1 
		  and exchange_date = to_date($2, 'yyyy-mm-dd')
	`

	err = tx.QueryRow(query, currencyID, date).Scan(&count)

	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		return err
	}

	if count == 0 {
		query = `
			insert into exchange_rate (
				currency_id,
				exchange_date,       
				rate
			)
			values (
				$1, to_date($2, 'yyyy-mm-dd'), $3
			)
		`

		_, err = tx.Exec(query, currencyID, date, exchRate/multiplier)

		if err != nil {
			return err
		}
	}

	return nil
}
