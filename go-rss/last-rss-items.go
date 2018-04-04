package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/geo-stanciu/go-utils/utils"
)

// RssLink - statistics for each rss link
type RssLink struct {
	Link        string
	RssDate     time.Time
	NewRssItems int
}

// SourceLastRSS - LastRSS for source
type SourceLastRSS struct {
	sync.RWMutex
	SourceID          int `sql:"rss_source_id"`
	LoweredSourceName string
	LastRssDate       time.Time `sql:"last_rss_date"`
	Links             []*RssLink
}

// GetLink - Get Rss link statistics by link
func (s *SourceLastRSS) GetLink(lnk string) *RssLink {
	for _, elem := range s.Links {
		if elem.Link == lnk {
			return elem
		}
	}

	return nil
}

// LastRssItems - Last Rss dates
type LastRssItems struct {
	sync.RWMutex
	RSS []*SourceLastRSS
}

// rssExists - Check if Rss exists in list
func (r *LastRssItems) rssExists(sourceID int) bool {
	for _, elem := range r.RSS {
		if elem.SourceID == sourceID {
			return true
		}
	}

	return false
}

// AddRSS - Add RSS elem
func (r *LastRssItems) AddRSS(s *SourceLastRSS) error {
	if s.SourceID > 0 && r.rssExists(s.SourceID) {
		return fmt.Errorf("element already exists")
	}

	r.RSS = append(r.RSS, s)

	return nil
}

// GetRSS - Get RSS elem
func (r *LastRssItems) GetRSS(sourceID int) *SourceLastRSS {
	for _, elem := range r.RSS {
		if elem.SourceID == sourceID {
			return elem
		}
	}

	return nil
}

// GetRSSBySource - Get RSS elem by source name
func (r *LastRssItems) GetRSSBySource(sourceName string) (*SourceLastRSS, error) {
	loweredSrcName := strings.ToLower(sourceName)

	for _, elem := range r.RSS {
		if elem.LoweredSourceName == loweredSrcName {
			return elem, nil
		}
	}

	return nil, nil
}

// SavelastDates - Save last RSS Dates
func (r *LastRssItems) SavelastDates() error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, elem := range r.RSS {
		// for rss's that have multiple links, I get the min last rss
		var minLastRSS time.Time
		epochStart, _ := utils.String2date("1970-01-01", utils.UTCDate)

		for _, lnk := range elem.Links {
			if minLastRSS.IsZero() ||
				minLastRSS.Equal(epochStart) ||
				(lnk.RssDate.After(epochStart) && minLastRSS.After(lnk.RssDate)) {

				minLastRSS = lnk.RssDate
			}
		}

		if !minLastRSS.IsZero() {
			pq := dbutl.PQuery(`
				UPDATE rss_source
				SET last_rss_date = ?
				WHERE rss_source_id = ?
				AND (last_rss_date IS NULL OR last_rss_date < ?)
			`, minLastRSS.UTC(),
				elem.SourceID,
				minLastRSS.UTC())

			_, err = dbutl.ExecTx(tx, pq)
			if err != nil {
				return err
			}
		}
	}

	tx.Commit()

	return nil
}
