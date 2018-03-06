package main

import (
	"database/sql"
	"encoding/xml"
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
	appName    = "RssGather"
	appVersion = "0.0.0.2"
	log        = logrus.New()
	audit      = utils.AuditLog{}
	db         *sql.DB
	dbutl      *utils.DbUtils
	config     = configuration{}
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.Formatter = new(logrus.JSONFormatter)
	log.Level = logrus.DebugLevel

	dbutl = new(utils.DbUtils)
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

	audit.SetLogger(appName+"/"+appVersion, log, dbutl)
	audit.SetWaitGroup(&wg)
	defer audit.Close()

	mw := io.MultiWriter(os.Stdout, audit)
	log.Out = mw

	for _, rss := range config.Rss {
		audit.Log(nil,
			"get rss",
			"save rss",
			"lang", rss.Lang,
			"source", rss.SourceName,
			"link", rss.Link)

		err = getStreamFromURL(&rss, parseXMLSource)
		if err != nil {
			log.Println(err)
			return
		}

		time.Sleep(50 * time.Millisecond)
	}

	audit.Log(nil, "gather rss", "Import done.")
	wg.Wait()
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
