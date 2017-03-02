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

var (
	config = Configuration{}
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

	cfgFile := "./conf.json"
	err = config.ReadFromFile(cfgFile)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	db, err := connect2Database(config.DbURL)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	err = db.Ping()

	if err != nil {
		log.Fatal(err)
	}

	err = prepareCurrencies(db)

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

func connect2Database(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		return nil, err
	}

	return db, nil
}

func prepareCurrencies(db *sql.DB) error {
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

			if rate.Multiplier == "-" {
				multiplier = 1.0
			} else if len(rate.Multiplier) > 0 {
				multiplier, err = strconv.ParseFloat(rate.Multiplier, 64)

				if err != nil {
					return err
				}
			}

			if rate.Rate == "-" {
				continue
			} else {
				exchRate, err = strconv.ParseFloat(rate.Rate, 64)

				if err != nil {
					return err
				}
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
