package main

import (
	"database/sql"
	"encoding/xml"
	"strings"
	"time"

	"github.com/geo-stanciu/go-utils/utils"
)

// RssEnclosure - RssEnclosure Item struct
type RssEnclosure struct {
	XMLName xml.Name `xml:"enclosure"`
	URL     string   `xml:"url,attr" sql:"enclosure_link"`
	Length  int      `xml:"length,attr" sql:"enclosure_length"`
	Type    string   `xml:"type,attr" sql:"enclosure_filetype"`
}

// RssMediaContent - RssEnclosure Item struct
type RssMediaContent struct {
	XMLName xml.Name `xml:"content"`
	URL     string   `xml:"url,attr" sql:"media_link"`
	Type    string   `xml:"type,attr" sql:"media_filetype"`
}

// RssItem - Rss Item struct
type RssItem struct {
	XMLName      xml.Name `xml:"item"`
	RssID        int64    `xml:"-" sql:"rss_id"`
	Title        string   `xml:"title" sql:"title"`
	Description  string   `xml:"description" sql:"description"`
	Link         string   `xml:"link" sql:"link"`
	ItemGUID     string   `xml:"guid" sql:"item_guid"`
	Date         string   `xml:"pubDate" sql:"sdate"`
	Category     string   `xml:"category"`
	SubCategory  string   `xml:"subcategory"`
	Content      string   `xml:"encoded"`
	Tags         string   `xml:"tags"`
	Creator      string   `xml:"creator"`
	MediaContent RssMediaContent
	Enclosure    RssEnclosure
	RssDate      time.Time `sql:"rss_date"`
}

// RssImage - RssImage Item struct
type RssImage struct {
	XMLName xml.Name `xml:"image"`
	Title   string   `xml:"title" sql:"image_title"`
	Width   string   `xml:"width" sql:"image_width"`
	Height  string   `xml:"height" sql:"image_height"`
	Link    string   `xml:"link" sql:"image_link"`
	URL     string   `xml:"url" sql:"image_url"`
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
	Generator   string    `xml:"generator"`
	WebMaster   string    `xml:"webMaster"`
	Copyright   string    `xml:"copyright"`
	Image       RssImage
	Rss         []*RssItem
}

func (r *RssFeed) getID(tx *sql.Tx) error {
	pq := dbutl.PQuery(`
		SELECT rss_source_id
		  FROM rss_source
		 WHERE lowered_source_name = ?
	`, strings.ToLower(r.Source))

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
				lowered_source_name,
				language,
				copyright,
				source_link,
				title,
				description,
				generator,
				web_master,
				image_title,
				image_width,
				image_heigth,
				image_link,
				image_url,
				add_date
			)
			VALUES (
				?, ?, ?, ?, ?,
				?, ?, ?, ?, ?,
				?, ?, ?, ?, ?
			)
		`, r.Source,
			strings.ToLower(r.Source),
			r.Language,
			r.Copyright,
			strings.TrimSpace(r.Link),
			strings.TrimSpace(r.Title),
			strings.TrimSpace(r.Description),
			strings.TrimSpace(r.Generator),
			strings.TrimSpace(r.WebMaster),
			strings.TrimSpace(r.Image.Title),
			r.Image.Width,
			r.Image.Height,
			strings.TrimSpace(r.Image.Link),
			strings.TrimSpace(r.Image.URL),
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
			       lowered_source_name = ?,
				   language = ?,
				   copyright = ?,
				   source_link = ?,
				   title = ?,
				   description = ?,
				   generator = ?,
				   web_master = ?,
				   image_title = ?,
				   image_width = ?,
				   image_heigth = ?,
				   image_link = ?,
				   image_url = ?
			 WHERE rss_source_id = ?
			   AND (
				   source_name IS NULL OR
				   source_name <> ? OR
				   lowered_source_name IS NULL OR
				   lowered_source_name <> ? OR
				   language IS NULL OR
				   language <> ? OR
				   copyright IS NULL OR
				   copyright <> ? OR
				   source_link IS NULL OR
				   source_link <> ? OR
				   title IS NULL OR
				   title <> ? OR
				   description IS NULL OR
				   description <> ? OR
				   generator IS NULL OR
				   generator <> ? OR
				   web_master IS NULL OR
				   web_master <> ? OR
				   image_title Is NULL OR
				   image_title <> ? OR
				   image_width IS NULL OR
				   image_width <> ? OR
				   image_heigth IS NULL OR
				   image_heigth <> ? OR
				   image_link IS NULL OR
				   image_link <> ? OR
				   image_url IS NULL OR
				   image_url <> ?
			   )
		`, r.Source,
			strings.ToLower(r.Source),
			r.Language,
			r.Copyright,
			strings.TrimSpace(r.Link),
			strings.TrimSpace(r.Title),
			strings.TrimSpace(r.Description),
			strings.TrimSpace(r.Generator),
			strings.TrimSpace(r.WebMaster),
			strings.TrimSpace(r.Image.Title),
			r.Image.Width,
			r.Image.Height,
			strings.TrimSpace(r.Image.Link),
			strings.TrimSpace(r.Image.URL),
			r.SourceID,
			r.Source,
			strings.ToLower(r.Source),
			r.Language,
			r.Copyright,
			strings.TrimSpace(r.Link),
			strings.TrimSpace(r.Title),
			strings.TrimSpace(r.Description),
			strings.TrimSpace(r.Generator),
			strings.TrimSpace(r.WebMaster),
			strings.TrimSpace(r.Image.Title),
			r.Image.Width,
			r.Image.Height,
			strings.TrimSpace(r.Image.Link),
			strings.TrimSpace(r.Image.URL))

		_, err = dbutl.ExecTx(tx, pq)
		if err != nil {
			return err
		}
	}

	if len(r.LastDate) > 0 {
		lastDate, err := utils.ParseRSSDate(r.LastDate)
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
			rss.RssDate, err = utils.ParseRSSDate(rss.Date)
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
				item_guid,
				category,
				subcategory,
				content,
				tags,
				creator,
				enclosure_link,
				enclosure_length,
				enclosure_filetype,
				media_link,
				media_filetype,
				rss_date,
				add_date
			)
			VALUES (
				?, ?, ?, ?, ?,
				?, ?, ?, ?, ?,
				?, ?, ?, ?, ?,
				?, ?
			)
		`, r.SourceID,
			strings.TrimSpace(rss.Title),
			strings.TrimSpace(rss.Link),
			strings.TrimSpace(rss.Description),
			strings.TrimSpace(rss.ItemGUID),
			strings.TrimSpace(rss.Category),
			strings.TrimSpace(rss.SubCategory),
			strings.TrimSpace(rss.Content),
			strings.TrimSpace(rss.Tags),
			strings.TrimSpace(rss.Creator),
			strings.TrimSpace(rss.Enclosure.URL),
			rss.Enclosure.Length,
			rss.Enclosure.Type,
			strings.TrimSpace(rss.MediaContent.URL),
			rss.MediaContent.Type,
			rss.RssDate.UTC(),
			dt)

		_, err = dbutl.ExecTx(tx, pq)
		if err != nil {
			return err
		}

		if config.CountNewRssItems {
			mutex.Lock()
			newRssItems++
			mutex.Unlock()
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
			   AND (last_rss_date IS NULL OR last_rss_date <> ?)
		`, lastRss.UTC(),
			r.SourceID,
			lastRss.UTC())

		_, err = dbutl.ExecTx(tx, pq)
		if err != nil {
			return err
		}
	}

	return err
}
