package main

import (
	"database/sql"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/geo-stanciu/go-utils/utils"
	"github.com/sirupsen/logrus"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	//_ "github.com/mattn/go-oci8"
)

var (
	appName     = "RssGather"
	appVersion  = "0.0.2.0"
	log         = logrus.New()
	audit       = utils.AuditLog{}
	db          *sql.DB
	dbutl       *utils.DbUtils
	config      = configuration{}
	queue       chan rssSource
	mutex       sync.RWMutex
	errFound    = false
	newRssItems = 0
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.Formatter = new(logrus.JSONFormatter)
	log.Level = logrus.DebugLevel

	dbutl = new(utils.DbUtils)

	queue = make(chan rssSource, 32)
}

// ParseSourceStream - Parse Source Stream
type ParseSourceStream func(rss *rssSource, source io.Reader) error

func main() {
	var err error
	var wg sync.WaitGroup

	cfgFile := "./conf.json"
	err = config.ReadFromFile(cfgFile)
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

	// initialize the rss readers
	for i := 0; i < config.RSSParalelReaders; i++ {
		go dealWithRSS(&wg)
	}

	for _, rss := range config.Rss {
		queue <- rss
	}

	done := rssSource{Done: true}
	for i := 0; i < config.RSSParalelReaders; i++ {
		wg.Add(1)
		queue <- done
	}

	// wait for all rss to be done
	wg.Wait()

	mutex.Lock()

	if errFound {
		err = errors.New("errors found while gathering rss")
		if config.CountNewRssItems {
			audit.Log(err, "gather rss", "Import failed.", "new_rss_items", newRssItems)
		} else {
			audit.Log(err, "gather rss", "Import failed.")
		}
	} else {
		if config.CountNewRssItems {
			audit.Log(nil, "gather rss", "Import done.", "new_rss_items", newRssItems)
		} else {
			audit.Log(nil, "gather rss", "Import done.")
		}
	}

	mutex.Unlock()

	// wait for all logs to be written
	wg.Wait()
}

func dealWithRSS(wg *sync.WaitGroup) {
	for {
		rss := <-queue

		if rss.Done {
			wg.Done()
			break
		}

		wg.Add(1)

		err := getStreamFromURL(&rss, parseXMLSource)

		if err != nil {
			mutex.Lock()
			errFound = true
			mutex.Unlock()
		}

		audit.Log(err,
			"get rss",
			"save rss",
			"lang", rss.Lang,
			"source", rss.SourceName,
			"link", rss.Link)

		time.Sleep(10 * time.Millisecond)
		wg.Done()
	}
}

func getStreamFromURL(rss *rssSource, callback ParseSourceStream) error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", rss.Link, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:58.0) Gecko/20100101 Firefox/58.0")

	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	err = callback(rss, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func parseXMLSource(rss *rssSource, source io.Reader) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	decoder := xml.NewDecoder(source)

	var feed RssFeed
	feed.Source = rss.SourceName
	feed.Link = rss.Link
	feed.Language = rss.Lang

	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}

		switch se := t.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "title":
				decoder.DecodeElement(&feed.Title, &se)
			case "description":
				decoder.DecodeElement(&feed.Description, &se)
			case "link":
				decoder.DecodeElement(&feed.Link, &se)
			case "language":
				decoder.DecodeElement(&feed.Language, &se)
			case "pubDate":
				decoder.DecodeElement(&feed.Date, &se)
			case "lastBuildDate":
				decoder.DecodeElement(&feed.LastDate, &se)
			case "generator":
				decoder.DecodeElement(&feed.Generator, &se)
			case "webMaster":
				decoder.DecodeElement(&feed.WebMaster, &se)
			case "copyright":
				decoder.DecodeElement(&feed.Copyright, &se)
			case "image":
				var img RssImage
				decoder.DecodeElement(&img, &se)
				feed.Image = img
			case "item":
				var item RssItem
				decoder.DecodeElement(&item, &se)
				feed.Rss = append(feed.Rss, &item)
			}
		}
	}

	err = feed.Save(tx)
	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}
