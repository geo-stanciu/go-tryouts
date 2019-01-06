package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/xml"
	"errors"
	"flag"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html/charset"

	"github.com/geo-stanciu/go-utils/utils"
	"github.com/sirupsen/logrus"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	//_ "github.com/mattn/go-oci8"
)

var (
	appName    = "RssGather"
	appVersion = "0.0.8.0"
	log        = logrus.New()
	audit      = utils.AuditLog{}
	db         *sql.DB
	dbutl      *utils.DbUtils
	config     = configuration{}
	queue      chan rssSource
	mutex      sync.RWMutex
	errFound   = false
	newItems   = 0
	lastFeeds  = LastFeeds{}
	rssLock    sync.RWMutex
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.Formatter = new(logrus.JSONFormatter)
	log.Level = logrus.DebugLevel

	dbutl = new(utils.DbUtils)

	queue = make(chan rssSource, 1024)
}

// ParseSourceStream - Parse Source Stream
type ParseSourceStream func(rss *rssSource, source io.Reader) error

func main() {
	var err error
	var wg sync.WaitGroup

	cfgPtr := flag.String("c", "conf.json", "config file")

	flag.Parse()

	err = config.ReadFromFile(*cfgPtr)
	if err != nil {
		log.Println(err)
		return
	}

	err = config.ReadFromFile("./rss.json")
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
		lastUpdate, err := getLastRSS(rss.SourceName)
		if err != nil {
			audit.Log(err, "gather rss", "Import failed.")
			return
		}
		rss.LastUpdate = lastUpdate

		if len(rss.Link) > 0 {
			queue <- rss
		}

		for _, lnk := range rss.Links {
			rss1 := rssSource{
				SourceName: rss.SourceName,
				Lang:       rss.Lang,
				Link:       lnk,
				LastUpdate: rss.LastUpdate,
			}
			queue <- rss1
		}
	}

	// wait for all rss to be done
	wg.Wait()

	err = lastFeeds.SavelastDates()
	if err != nil {
		audit.Log(err, "gather rss", "Import failed.")
	} else {
		mutex.Lock()

		if errFound {
			err = errors.New("errors found while gathering rss")
			if config.CountNewRssItems {
				audit.Log(err, "gather rss", "Import failed.", "new_rss_items", newItems)
			} else {
				audit.Log(err, "gather rss", "Import failed.")
			}
		} else {
			if config.CountNewRssItems {
				audit.Log(nil, "gather rss", "Import done.", "new_rss_items", newItems)
			} else {
				audit.Log(nil, "gather rss", "Import done.")
			}
		}

		mutex.Unlock()
	}

	// wait for all logs to be written
	wg.Wait()

	close(queue)
}

func dealWithRSS(wg *sync.WaitGroup) {
	for {
		rss, ok := <-queue
		if !ok {
			break
		}

		startTime := time.Now()

		wg.Add(1)

		var err error

		lastFeeds.Lock()
		rss.Feed, err = lastFeeds.GetFeedBySource(rss.SourceName)
		if err != nil {
			endTime := time.Now()

			audit.Log(err,
				"get rss",
				"save rss",
				"lang", rss.Lang,
				"source", rss.SourceName,
				"link", rss.Link,
				"new_rss_items", 0,
				"time_elapsed_ms", endTime.Sub(startTime)/1E6,
			)

			lastFeeds.Unlock()
			wg.Done()
			continue
		}

		if rss.Feed == nil {
			rss.Feed = new(RSSFeed)
			rss.Feed.Initialize(rss.SourceName)
			lastFeeds.AddRSS(rss.Feed)
		}

		if rss.FeedLnk = rss.Feed.GetLink(rss.Link); rss.FeedLnk == nil {
			rss.FeedLnk = new(RssLink)
			rss.FeedLnk.Link = rss.Link
			rss.Feed.Links = append(rss.Feed.Links, rss.FeedLnk)
		}
		lastFeeds.Unlock()

		err = getStreamFromURL(&rss, parseXMLSource)

		if err != nil {
			mutex.Lock()
			errFound = true
			mutex.Unlock()
		}

		newRss := 0
		rss.Feed.Lock()
		newRss = rss.FeedLnk.NewItems
		rss.Feed.Unlock()

		endTime := time.Now()

		audit.Log(err,
			"get rss",
			"save rss",
			"lang", rss.Lang,
			"source", rss.SourceName,
			"link", rss.Link,
			"new_rss_items", newRss,
			"time_elapsed_ms", endTime.Sub(startTime)/1E6,
		)

		time.Sleep(10 * time.Millisecond)

		wg.Done()
	}
}

func getStreamFromURL(rss *rssSource, callback ParseSourceStream) error {
	var client *http.Client

	if rss.TrustCert && strings.HasPrefix(rss.Link, "https") {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		client = &http.Client{Transport: tr}
	} else {
		client = &http.Client{}
	}

	req, err := http.NewRequest("GET", rss.Link, nil)
	if err != nil {
		return err
	}

	// NOTE this !! - close the request
	req.Close = true

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:64.0) Gecko/20100101 Firefox/64.0")

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
	decoder := xml.NewDecoder(source)
	decoder.CharsetReader = charset.NewReaderLabel

	var feed RssFeed
	feed.Source = rss.SourceName
	feed.Language = rss.Lang
	feed.Link = rss.Link
	feed.Feed = rss.Feed
	feed.FeedLink = rss.FeedLnk

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
			switch se.Name.Local {
			case "title":
				decoder.DecodeElement(&feed.Title, &se)
			case "description":
				decoder.DecodeElement(&feed.Description, &se)
			case "language":
				decoder.DecodeElement(&feed.Language, &se)
			case "link":
				decoder.DecodeElement(&feed.Link, &se)
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

	rssLock.Lock()
	defer rssLock.Unlock()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = feed.Save(tx)
	if err != nil {
		return err
	}

	tx.Commit()

	return nil
}
