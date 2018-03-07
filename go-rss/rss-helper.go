package main

import (
	"database/sql"
	"encoding/xml"
	"time"

	"github.com/geo-stanciu/go-utils/utils"
)

// RssEnclosure - RssEnclosure Item struct
type RssEnclosure struct {
	XMLName xml.Name `xml:"enclosure"`
	URL     string   `xml:"url,attr"`
	Length  int      `xml:"length,attr"`
	Type    string   `xml:"type,attr"`
}

// RssItem - Rss Item struct
type RssItem struct {
	XMLName     xml.Name `xml:"item"`
	RssID       int64    `xml:"-" sql:"rss_id"`
	Title       string   `xml:"title" sql:"title"`
	Description string   `xml:"description" sql:"description"`
	Link        string   `xml:"link" sql:"link"`
	Date        string   `xml:"pubDate" sql:"sdate"`
	Category    string   `xml:"category"`
	Enclosure   RssEnclosure
	RssDate     time.Time `sql:"rss_date"`
}

// RssFeed - Rss Channel struct
type RssFeed struct {
	XMLName     xml.Name  `xml:"channel"`
	SourceID    int       `xml:"-" sql:"rss_source_id"`
	Source      string    `sql:"source_name"`
	Title       string    `xml:"title" sql:"source_title"`
	Description string    `xml:"description" sql:"source_description"`
	Link        string    `xml:"link" sql:"source_title"`
	Language    string    `xml:"language" sql:"language"`
	Date        string    `xml:"pubDate"`
	LastDate    string    `xml:"lastBuildDate"`
	LastRssDate time.Time `sql:"last_rss_date"`
	Rss         []*RssItem
}

func (r *RssFeed) getID(tx *sql.Tx) error {
	pq := dbutl.PQuery(`
		SELECT rss_source_id
		  FROM rss_source
		 WHERE lower(source_name) = lower(?)
	`, r.Source)

	err := tx.QueryRow(pq.Query, pq.Args...).Scan(&r.SourceID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	return nil
}

// Save - save rss news
func (r *RssFeed) Save(tx *sql.Tx) error {
	var pq *utils.PreparedQuery
	var err error

	dt := time.Now().UTC()

	epochStart, err := utils.String2date("1970-01-01", utils.UTCDate)
	if err != nil {
		return err
	}

	if r.SourceID <= 0 {
		r.getID(tx)
		if err != nil {
			return err
		}
	}

	if r.SourceID <= 0 {
		pq = dbutl.PQuery(`
			INSERT INTO rss_source (
				source_name,
				language,
				source_link,
				title,
				description,
				add_date
			)
			VALUES (
				?, ?, ?, ?, ?, ?
			)
		`, r.Source,
			r.Language,
			r.Link,
			r.Title,
			r.Description,
			dt)

		_, err = dbutl.ExecTx(tx, pq)
		if err != nil {
			return err
		}

		r.getID(tx)
		if err != nil {
			return err
		}

		r.LastRssDate = epochStart
	} else {
		pq = dbutl.PQuery(`
			SELECT CASE
					 WHEN last_rss_date IS NULL THEN
					   DATE ?
					 ELSE
					   last_rss_date
					 END last_rss_date
			  FROM rss_source
			 WHERE rss_source_id = ?
		`, "1970-01-01",
			r.SourceID)

		err := tx.QueryRow(pq.Query, pq.Args...).Scan(&r.LastRssDate)
		if err != nil {
			return err
		}

		pq = dbutl.PQuery(`
			UPDATE rss_source
			   SET source_name = ?,
				   language = ?,
				   source_link = ?,
				   title = ?,
				   description = ?
			 WHERE rss_source_id = ?
			   AND (
			       source_name != ? OR
				   language != ? OR
				   source_link != ? OR
				   title != ? OR
				   description != ?
			   )
		`, r.Source,
			r.Language,
			r.Link,
			r.Title,
			r.Description,
			r.SourceID,
			r.Source,
			r.Language,
			r.Link,
			r.Title,
			r.Description)

		_, err = dbutl.ExecTx(tx, pq)
		if err != nil {
			return err
		}
	}

	if len(r.LastDate) > 0 {
		lastDate, err := parseRSSDate(r.LastDate)
		if err != nil {
			return err
		}

		if !lastDate.After(r.LastRssDate) {
			return nil
		}
	}

	lastRss := epochStart

	for _, rss := range r.Rss {
		if len(rss.Date) == 0 {
			found := false

			pq = dbutl.PQuery(`
				SELECT CASE WHEN EXISTS (
					SELECT 1
					  FROM rss
					 WHERE rss_source_id = ?
					   AND title = ?
					   AND link = ?
				) THEN 1 ELSE 0 END
				FROM dual
			`, r.SourceID,
				rss.Title,
				rss.Link)

			err = tx.QueryRow(pq.Query, pq.Args...).Scan(&found)
			if err != nil {
				return err
			}

			if found {
				continue
			}

			rss.RssDate = time.Now().UTC()
		} else {
			rss.RssDate, err = parseRSSDate(rss.Date)
			if err != nil {
				return err
			}
		}

		if !rss.RssDate.After(r.LastRssDate) {
			continue
		}

		pq = dbutl.PQuery(`
			INSERT INTO rss (
				rss_source_id,
				title,
				link,
				description,
				category,
				enclosure_link,
				enclosure_length,
				enclosure_filetype,
				rss_date,
				add_date
			)
			VALUES (
				?, ?, ?, ?, ?, ?, ?, ?, ?, ?
			)
		`, r.SourceID,
			rss.Title,
			rss.Link,
			rss.Description,
			rss.Category,
			rss.Enclosure.URL,
			rss.Enclosure.Length,
			rss.Enclosure.Type,
			rss.RssDate.UTC(),
			dt)

		_, err = dbutl.ExecTx(tx, pq)
		if err != nil {
			return err
		}

		if lastRss.Before(rss.RssDate) {
			lastRss = rss.RssDate
		}
	}

	if lastRss.After(epochStart) {
		r.LastRssDate = lastRss

		pq = dbutl.PQuery(`
			UPDATE rss_source
			   SET last_rss_date = ?
			 WHERE rss_source_id = ?
		`, lastRss.UTC(),
			r.SourceID)

		_, err = dbutl.ExecTx(tx, pq)
		if err != nil {
			return err
		}
	}

	return err
}

func parseRSSDate(sdate string) (time.Time, error) {
	var err1, err2 error
	var dt time.Time
	dt, err1 = utils.String2date(sdate, utils.RSSDateTimeTZ)
	if err1 != nil {
		dt, err2 = utils.String2date(sdate, utils.RSSDateTime)
		if err2 != nil {
			return dt.UTC(), err1
		}
	}

	return dt.UTC(), nil
}
