package main

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
)

var (
	log    = logrus.New()
	db     *sql.DB
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

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.Formatter = new(logrus.JSONFormatter)
	log.Level = logrus.DebugLevel
}

func main() {
	var auditLog AuditLog
	var xmlBytes []byte
	var err error

	cfgFile := "./conf.json"
	err = config.ReadFromFile(cfgFile)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	err = connect2Database(config.DbURL)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer db.Close()

	mw := io.MultiWriter(os.Stdout, auditLog)
	log.Out = mw

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

func connect2Database(dbURL string) error {
	var err error

	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("Can't connect to the database, error: %s", err.Error())
	}

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("Can't ping the database, error: %s", err.Error())
	}

	return nil
}

func prepareCurrencies(db *sql.DB) error {
	var found bool

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := "SELECT EXISTS(SELECT 1 FROM currency)"

	err = tx.QueryRow(query).Scan(&found)

	switch {
	case err == sql.ErrNoRows:
		return err
	case err != nil:
		return err
	}

	if !found {
		query = "INSERT INTO currency (currency) VALUES ($1)"

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
		log.WithField("data", cube.Date).Info("Importing exchange rates...")

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

	log.Info("Import done.")

	return nil
}

func storeRate(tx *sql.Tx, date string, currency string, multiplier float64, exchRate float64) error {
	var currencyID int32
	var found bool

	query := "SELECT currency_id FROM currency WHERE currency = $1"

	err := tx.QueryRow(query, currency).Scan(&currencyID)

	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		return err
	}

	query = `
		SELECT EXISTS(
			SELECT 1
			FROM exchange_rate 
		   WHERE currency_id = $1 
		     AND exchange_date = to_date($2, 'yyyy-mm-dd')
		)
	`

	err = tx.QueryRow(query, currencyID, date).Scan(&found)

	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		return err
	}

	if !found {
		query = `
			INSERT INTO exchange_rate (
				currency_id,
				exchange_date,       
				rate
			)
			VALUES (
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
