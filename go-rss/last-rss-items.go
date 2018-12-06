package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/geo-stanciu/go-utils/utils"
)

// LastFeeds - Last Rss dates
type LastFeeds struct {
	sync.RWMutex
	RSS []*RSSFeed
}

// rssExists - Check if Rss exists in list
func (r *LastFeeds) rssExists(sourceID int) bool {
	for _, elem := range r.RSS {
		if elem.SourceID == sourceID {
			return true
		}
	}

	return false
}

// AddRSS - Add RSS elem
func (r *LastFeeds) AddRSS(s *RSSFeed) error {
	if s.SourceID > 0 && r.rssExists(s.SourceID) {
		return fmt.Errorf("element already exists")
	}

	r.RSS = append(r.RSS, s)

	return nil
}

// GetRSS - Get RSS elem
func (r *LastFeeds) GetRSS(sourceID int) *RSSFeed {
	for _, elem := range r.RSS {
		if elem.SourceID == sourceID {
			return elem
		}
	}

	return nil
}

// GetFeedBySource - Get RSS elem by source name
func (r *LastFeeds) GetFeedBySource(sourceName string) (*RSSFeed, error) {
	loweredSrcName := strings.ToLower(sourceName)

	for _, elem := range r.RSS {
		if elem.LoweredSourceName == loweredSrcName {
			return elem, nil
		}
	}

	return nil, nil
}

// SavelastDates - Save last RSS Dates
func (r *LastFeeds) SavelastDates() error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err = dbutl.SetAsyncCommit(tx); err != nil {
		return err
	}

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

		var pq *utils.PreparedQuery

		if !minLastRSS.IsZero() {
			pq = dbutl.PQuery(`
				UPDATE rss_source
				   SET last_rss_date = ?
				 WHERE rss_source_id = ?
				   AND (last_rss_date IS NULL OR last_rss_date < ?)
			`, minLastRSS.UTC(),
				elem.SourceID,
				minLastRSS.UTC())
		} else {
			pq = dbutl.PQuery(`
				UPDATE rss_source
				   SET last_rss_date = ?
				 WHERE rss_source_id = ?
				   AND last_rss_date IS NULL
			`, "1970-01-01 00:00:00",
				elem.SourceID)
		}

		_, err = dbutl.ExecTx(tx, pq)
		if err != nil {
			return err
		}
	}

	tx.Commit()

	return nil
}
