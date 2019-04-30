package main

import (
	"database/sql"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/geo-stanciu/go-utils/utils"
	"github.com/sirupsen/logrus"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-oci8"
	_ "github.com/mattn/go-sqlite3"
)

var (
	appName    = "GoExchRates"
	appVersion = "0.0.4.0"
	log        = logrus.New()
	audit      = utils.AuditLog{}
	db         *sql.DB
	dbutl      *utils.DbUtils
	config     = configuration{}
	currentDir string
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.Formatter = new(logrus.JSONFormatter)
	log.Level = logrus.DebugLevel

	dbutl = new(utils.DbUtils)
	currentDir = filepath.Dir(os.Args[0])
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

	cfgPtr := flag.String("c", fmt.Sprintf("%s/conf.json", currentDir), "config file")

	flag.Parse()

	if _, err = os.Stat(*cfgPtr); os.IsNotExist(err) {
		err = config.ReadFromFile(fmt.Sprintf("%s/%s", currentDir, *cfgPtr))
	} else {
		err = config.ReadFromFile(*cfgPtr)
	}

	if err != nil {
		log.Println(err)
		return
	}

	err = dbutl.Connect2Database(&db, config.DbType, config.DbURL)
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()

	audit.SetLogger(appName, appVersion, log, dbutl)
	audit.SetWaitGroup(&wg)
	defer audit.Close()

	mw := io.MultiWriter(os.Stdout, audit)
	log.Out = mw

	err = prepareCurrencies()
	if err != nil {
		log.Println(err)
		return
	}

	err = getStreamFromURL(config.RatesXMLUrl, parseXMLSource)
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
	decoder := xml.NewDecoder(source)

	for {
		t, err := decoder.Token()
		if t == nil {
			break
		}
		if err != nil && err != io.EOF {
			return err
		}

		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "Cube" {
				var cube Cube
				decoder.DecodeElement(&cube, &se)

				if err := dealWithRates(&cube); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func dealWithRates(cube *Cube) error {
	tx, err := dbutl.BeginTransaction()
	if err != nil {
		return err
	}
	defer dbutl.Rollback(tx)

	if err = dbutl.SetAsyncCommit(tx); err != nil {
		tx.Rollback()
		return err
	}

	if err = storeRates(tx, *cube); err != nil {
		return err
	}

	dbutl.Commit(tx)

	return nil
}

func getCurrencyIfExists(tx *sql.Tx, currency string) (int32, error) {
	var currencyID int32

	pq := dbutl.PQuery(`
		SELECT currency_id FROM currency WHERE currency = ?
	`, currency)

	err := tx.QueryRow(pq.Query, pq.Args...).Scan(&currencyID)
	if err != nil && err != sql.ErrNoRows {
		return -1, err
	}

	return currencyID, nil
}

func addCurrencyIfNotExists(tx *sql.Tx, currency string) (int32, error) {
	currencyID, err := getCurrencyIfExists(tx, currency)
	if err != nil {
		return -1, err
	} else if currencyID > 0 {
		return currencyID, nil
	}

	pq := dbutl.PQuery(`
		INSERT INTO currency (currency) VALUES (?)
	`, currency)

	_, err = dbutl.ExecTx(tx, pq)
	if err != nil {
		return -1, err
	}

	audit.Log(nil, "add currency", "Adding missing currency...", "currency", currency)

	return addCurrencyIfNotExists(tx, currency)
}

func prepareCurrencies() error {
	tx, err := dbutl.BeginTransaction()
	if err != nil {
		return err
	}
	defer dbutl.Rollback(tx)

	if err = dbutl.SetAsyncCommit(tx); err != nil {
		return err
	}

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

	dbutl.Commit(tx)

	return nil
}

func storeRates(tx *sql.Tx, cube Cube) error {
	var err error
	var refCurrencyID int32

	audit.Log(nil, "exchange rates", "Importing exchange rates...", "date", cube.Date)

	refCurrencyID, err = addCurrencyIfNotExists(tx, "RON")
	if err != nil {
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
	found := 0
	var currencyID int32
	var err error

	exch := big.NewFloat(exchRate)
	mul := big.NewFloat(multiplier)
	srate := new(big.Float).SetMode(big.ToNearestAway).Quo(exch, mul).Text('f', 6)

	rate, err := strconv.ParseFloat(srate, 64)
	if err != nil {
		return err
	}

	if config.AddMissingCurrencies {
		currencyID, err = addCurrencyIfNotExists(tx, currency)
	} else {
		currencyID, err = getCurrencyIfExists(tx, currency)
	}

	if err != nil {
		return err
	}

	pq := dbutl.PQuery(`
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

	if found == 0 {
		pq = dbutl.PQuery(`
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

		_, err = dbutl.ExecTx(tx, pq)
		if err != nil {
			return err
		}

		audit.Log(nil,
			"add exchange rate",
			"added value",
			"date", date,
			"currency", currency,
			"rate", rate)
	}

	return nil
}
