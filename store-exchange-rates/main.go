package main

import (
	"database/sql"
	"encoding/xml"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/geo-stanciu/go-utils/utils"
	"github.com/sirupsen/logrus"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

var (
	log    = logrus.New()
	audit   = utils.AuditLog{}
	db     *sql.DB
	dbUtils = utils.DbUtils{}
	config = Configuration{}
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.Formatter = new(logrus.JSONFormatter)
	log.Level = logrus.DebugLevel
}

type Rate struct {
	Currency   string `xml:"currency,attr"`
	Multiplier string `xml:"multiplier,attr"`
	Rate       string `xml:",chardata"`
}

type Cube struct {
	Date string `xml:"date,attr"`
	Rate []Rate
}

type ParseSourceStream func(source io.Reader) error

func main() {
	var err error

	cfgFile := "./conf.json"
	err = config.ReadFromFile(cfgFile)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	err = dbUtils.Connect2Database(&db, config.DbType, config.DbURL)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer db.Close()

	audit.SetLoggerAndDatabase(log, &dbUtils)

	mw := io.MultiWriter(os.Stdout, audit)
	log.Out = mw

	err = prepareCurrencies()
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) >= 2 {
		err = getStreamFromFile(os.Args[1], parseXmlSource)
	} else {
		err = getStreamFromURL(config.RatesXMLUrl, parseXmlSource)
	}

	if err != nil {
		log.Fatal(err)
		return
	}

	log.Info("Import done.")
}

func getStreamFromURL(url string, callback ParseSourceStream) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	err = callback(response.Body)
	if err != nil {
		return err
	}

	return nil
}

func getStreamFromFile(filename string, callback ParseSourceStream) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	err = callback(f)
	if err != nil {
		return err
	}

	return nil
}

func parseXmlSource(source io.Reader) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	decoder := xml.NewDecoder(source)

	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}

		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "Cube" {
				var cube Cube
				decoder.DecodeElement(&cube, &se)

				err := storeRates(tx, cube)
				if err != nil {
					return err
				}
			}
		}
	}

	tx.Commit()

	return nil
}

func prepareCurrencies() error {
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
		query = dbUtils.PQuery("INSERT INTO currency (currency) VALUES (?)")

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

func storeRates(tx *sql.Tx, cube Cube) error {
	var err error

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

	return nil
}

func storeRate(tx *sql.Tx, date string, currency string, multiplier float64, exchRate float64) error {
	var currencyID int32
	var found bool

	query := dbUtils.PQuery("SELECT currency_id FROM currency WHERE currency = ?")

	err := tx.QueryRow(query, currency).Scan(&currencyID)

	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		return err
	}

	dData, err := string2date(date, ISODate)
	if err != nil {
		return err
	}

	query = dbUtils.PQuery(`
		SELECT EXISTS(
			SELECT 1
			FROM exchange_rate 
		   WHERE currency_id = ? 
		     AND exchange_date = ?
		)
	`)

	err = tx.QueryRow(query, currencyID, dData).Scan(&found)

	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		return err
	}

	if !found {
		query = dbUtils.PQuery(`
			INSERT INTO exchange_rate (
				currency_id,
				exchange_date,
				rate
			)
			VALUES (?, ?, ?)
		`)

		_, err = tx.Exec(query, currencyID, dData, exchRate/multiplier)
		if err != nil {
			return err
		}
	}

	return nil
}
