package main

import (
	"database/sql"
	"encoding/xml"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/geo-stanciu/go-utils/utils"
	"github.com/sirupsen/logrus"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	//_ "github.com/mattn/go-oci8"
)

var (
	appName    = "GoExchRates"
	appVersion = "0.0.0.2"
	log        = logrus.New()
	audit      = utils.AuditLog{}
	db         *sql.DB
	dbUtils    *utils.DbUtils
	config     = Configuration{}
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.Formatter = new(logrus.JSONFormatter)
	log.Level = logrus.DebugLevel

	dbUtils = new(utils.DbUtils)
}

// Rate - Exchange rate struct
type Rate struct {
	Currency   string `xml:"currency,attr"`
	Multiplier string `xml:"multiplier,attr"`
	Rate       string `xml:",chardata"`
}

// Cube - colection of exchange rates
type Cube struct {
	Date string `xml:"date,attr"`
	Rate []Rate
}

// ParseSourceStream - Parse Source Stream
type ParseSourceStream func(source io.Reader) error

func main() {
	var err error
	var wg sync.WaitGroup

	cfgFile := "./conf.json"
	err = config.ReadFromFile(cfgFile)
	if err != nil {
		log.Println(err)
		return
	}

	err = dbUtils.Connect2Database(&db, config.DbType, config.DbURL)
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()

	audit.SetLogger(appName+"/"+appVersion, log, dbUtils)
	audit.SetWaitGroup(&wg)
	defer audit.Close()

	mw := io.MultiWriter(os.Stdout, audit)
	log.Out = mw

	err = prepareCurrencies()
	if err != nil {
		log.Println(err)
		return
	}

	if len(os.Args) >= 2 {
		err = getStreamFromFile(os.Args[1], parseXMLSource)
	} else {
		err = getStreamFromURL(config.RatesXMLUrl, parseXMLSource)
	}

	if err != nil {
		log.Println(err)
		return
	}

	audit.Log(nil, "import exchange rates", "Import done.")
	wg.Wait()
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

func parseXMLSource(source io.Reader) error {
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

func addCurrencyIfNotExists(tx *sql.Tx, currency string) (int32, error) {
	var currencyID int32

	pq := dbUtils.PQuery(`
		SELECT currency_id FROM currency WHERE currency = ?
	`, currency)

	err := tx.QueryRow(pq.Query, pq.Args...).Scan(&currencyID)
	if err != nil && err != sql.ErrNoRows {
		return -1, err
	} else if currencyID > 0 {
		return currencyID, nil
	}

	pq = dbUtils.PQuery(`
		INSERT INTO currency (currency) VALUES (?)
	`, currency)

	_, err = dbUtils.ExecTx(tx, pq)
	if err != nil {
		return -1, err
	}

	return addCurrencyIfNotExists(tx, currency)
}

func prepareCurrencies() error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	refCurrencyID, err := addCurrencyIfNotExists(tx, "RON")
	if err != nil {
		return err
	}

	_, err = addCurrencyIfNotExists(tx, "EUR")
	if err != nil {
		return err
	}

	_, err = addCurrencyIfNotExists(tx, "USD")
	if err != nil {
		return err
	}

	_, err = addCurrencyIfNotExists(tx, "CHF")
	if err != nil {
		return err
	}

	err = storeRate(tx, "1970-01-01", refCurrencyID, "RON", 1.0, 1.0)
	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}

func storeRates(tx *sql.Tx, cube Cube) error {
	var err error
	var refCurrencyID int32

	audit.Log(nil, "import exchange rates", "Importing exchange rates...", "data", cube.Date)

	pq := dbUtils.PQuery(`
		SELECT currency_id FROM currency WHERE currency = ?
	`, "RON")

	err = tx.QueryRow(pq.Query, pq.Args...).Scan(&refCurrencyID)

	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		return err
	}

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

		err = storeRate(tx, cube.Date, refCurrencyID, rate.Currency, multiplier, exchRate)
		if err != nil {
			return err
		}
	}

	return nil
}

func storeRate(tx *sql.Tx, date string, refCurrencyID int32, currency string, multiplier float64, exchRate float64) error {
	var found bool
	var currencyID int32
	var err error
	rate := exchRate / multiplier

	if config.AddMissingCurrencies {
		currencyID, err = addCurrencyIfNotExists(tx, currency)
	} else {
		// get
	}

	if err != nil {
		return err
	}

	pq := dbUtils.PQuery(`
		SELECT CASE WHEN EXISTS (
			SELECT 1
			FROM exchange_rate 
		   WHERE currency_id = ? 
		     AND exchange_date = DATE ?
		) THEN 1 ELSE 0 END
		FROM dual
	`, currencyID,
		date)

	err = tx.QueryRow(pq.Query, pq.Args...).Scan(&found)
	if err != nil {
		return err
	}

	if !found {
		pq = dbUtils.PQuery(`
			INSERT INTO exchange_rate (
				reference_currency_id,
				currency_id,
				exchange_date,
				rate
			)
			VALUES (?, ?, DATE ?, ?)
		`, refCurrencyID,
			currencyID,
			date,
			rate)

		_, err = dbUtils.ExecTx(tx, pq)
		if err != nil {
			return err
		}

		audit.Log(nil,
			"import exchange rates",
			"add exchange rate",
			"data", date,
			"currency", currency,
			"rate", rate)
	}

	return nil
}
