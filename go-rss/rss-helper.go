package main

import (
	"database/sql"
	"encoding/xml"
	"strings"
	"sync"
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

// RssMediaThumbnail - RssMediaThumbnail Item struct
type RssMediaThumbnail struct {
	XMLName xml.Name `xml:"thumbnail"`
	URL     string   `xml:"url,attr" sql:"media_thumbnail"`
}

// RssItem - Rss Item struct
type RssItem struct {
	XMLName      xml.Name `xml:"item"`
	RssID        int64    `xml:"-" sql:"rss_id"`
	Title        string   `xml:"title" sql:"title"`
	Description  string   `xml:"description" sql:"description"`
	Link         string   `xml:"link" sql:"link"`
	ItemGUID     string   `xml:"guid" sql:"item_guid"`
	OrigLink     string   `xml:"origLink"`
	Date         string   `xml:"pubDate" sql:"sdate"`
	Keywords     string   `xml:"keywords"`
	Category     string   `xml:"category"`
	SubCategory  string   `xml:"subcategory"`
	Content      string   `xml:"encoded"`
	Tags         string   `xml:"tags"`
	Creator      string   `xml:"creator"`
	Thumbnail    RssMediaThumbnail
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
	XMLName     xml.Name `xml:"channel"`
	SourceID    int      `xml:"-" sql:"rss_source_id"`
	Source      string   `sql:"source_name"`
	Title       string   `xml:"title" sql:"source_title"`
	Description string   `xml:"description" sql:"source_description"`
	Link        string   `xml:"link" sql:"source_title"`
	Language    string   `xml:"language" sql:"language"`
	Date        string   `xml:"pubDate"`
	LastDate    string   `xml:"lastBuildDate"`
	Generator   string   `xml:"generator"`
	WebMaster   string   `xml:"webMaster"`
	Copyright   string   `xml:"copyright"`
	Image       RssImage
	SrcLastRss  *SourceLastRSS
	SrcRssLink  *RssLink
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

var newSourceLock sync.RWMutex

// Save - save rss news
func (r *RssFeed) Save(tx *sql.Tx) error {
	var pq *utils.PreparedQuery
	var err error

	dt := time.Now().UTC()

	epochStart, err := utils.String2date("1970-01-01", utils.UTCDate)
	if err != nil {
		return err
	}

	newSourceLock.Lock()

	if r.SourceID <= 0 {
		r.getID(tx)
		if err != nil {
			newSourceLock.Unlock()
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
			newSourceLock.Unlock()
			return err
		}

		r.getID(tx)
		if err != nil {
			newSourceLock.Unlock()
			return err
		}

		newSourceLock.Unlock()
	} else {
		newSourceLock.Unlock()

		r.SrcLastRss.Lock()
		if r.SrcLastRss.LastRssDate.IsZero() || r.SrcLastRss.LastRssDate.Equal(epochStart) {
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

			err := tx.QueryRow(pq.Query, pq.Args...).Scan(&r.SrcLastRss.LastRssDate)
			if err != nil {
				r.SrcLastRss.Unlock()
				return err
			}
		}
		r.SrcLastRss.Unlock()

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

		r.SrcLastRss.Lock()
		if !lastDate.After(r.SrcLastRss.LastRssDate) {
			r.SrcLastRss.Unlock()
			return nil
		}
		r.SrcLastRss.Unlock()
	}

	lastRssDate := epochStart

	for _, rss := range r.Rss {
		if len(rss.Date) == 0 {
			rss.RssDate = time.Now().UTC()
		} else {
			rss.RssDate, err = utils.ParseRSSDate(rss.Date)
			if err != nil {
				return err
			}
		}

		r.SrcLastRss.Lock()
		if !rss.RssDate.After(r.SrcLastRss.LastRssDate) {
			r.SrcLastRss.Unlock()
			continue
		}
		r.SrcLastRss.Unlock()

		found, err := r.rssExists(tx, rss.Title, rss.Link)
		if err != nil {
			return err
		}

		if found {
			continue
		}

		pq = dbutl.PQuery(`
			INSERT INTO rss (
				rss_source_id,
				title,
				link,
				description,
				item_guid,
				orig_link,
				category,
				subcategory,
				content,
				keywords,
				tags,
				creator,
				enclosure_link,
				enclosure_length,
				enclosure_filetype,
				media_link,
				media_filetype,
				media_thumbnail,
				rss_date,
				add_date
			)
			VALUES (
				?, ?, ?, ?, ?,
				?, ?, ?, ?, ?,
				?, ?, ?, ?, ?,
				?, ?, ?, ?, ?
			)
		`, r.SourceID,
			strings.TrimSpace(rss.Title),
			strings.TrimSpace(rss.Link),
			strings.TrimSpace(rss.Description),
			strings.TrimSpace(rss.ItemGUID),
			strings.TrimSpace(rss.OrigLink),
			strings.TrimSpace(rss.Category),
			strings.TrimSpace(rss.SubCategory),
			strings.TrimSpace(rss.Content),
			strings.TrimSpace(rss.Keywords),
			strings.TrimSpace(rss.Tags),
			strings.TrimSpace(rss.Creator),
			strings.TrimSpace(rss.Enclosure.URL),
			rss.Enclosure.Length,
			rss.Enclosure.Type,
			strings.TrimSpace(rss.MediaContent.URL),
			rss.MediaContent.Type,
			strings.TrimSpace(rss.Thumbnail.URL),
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

			r.SrcLastRss.Lock()
			r.SrcRssLink.NewRssItems++
			r.SrcLastRss.Unlock()
		}

		if lastRssDate.Before(rss.RssDate) {
			lastRssDate = rss.RssDate
		}
	}

	r.SrcLastRss.Lock()
	if r.SrcRssLink.RssDate.IsZero() || r.SrcRssLink.RssDate.Before(lastRssDate) {
		r.SrcRssLink.RssDate = lastRssDate
	}
	r.SrcLastRss.Unlock()

	return err
}

func (r *RssFeed) rssExists(tx *sql.Tx, title string, link string) (bool, error) {
	found := false

	pq := dbutl.PQuery(`
		SELECT CASE WHEN EXISTS (
			SELECT 1
			  FROM rss
			 WHERE rss_source_id = ?
			   AND title = ?
			   AND link = ?
		) THEN 1 ELSE 0 END
		FROM dual
	`, r.SourceID,
		strings.TrimSpace(title),
		strings.TrimSpace(link))

	err := tx.QueryRow(pq.Query, pq.Args...).Scan(&found)
	if err != nil {
		return false, err
	}

	return found, nil
}

func getLastRSS(source string) (time.Time, error) {
	var err error
	var lastRSS time.Time

	pq := dbutl.PQuery(`
		SELECT CASE
			     WHEN last_rss_date IS NULL THEN
				   DATE ?
				 ELSE
				   last_rss_date
				END last_rss_date
			  FROM rss_source
		 WHERE lowered_source_name = ?
	`, "1970-01-01",
		strings.ToLower(source))

	err = db.QueryRow(pq.Query, pq.Args...).Scan(&lastRSS)
	if err != nil {
		if err == sql.ErrNoRows {
			epochStart, err := utils.String2date("1970-01-01", utils.UTCDate)
			if err != nil {
				return time.Now(), err
			}

			lastRSS = epochStart
		} else {
			return time.Now(), err
		}
	}

	return lastRSS.UTC(), nil
}
