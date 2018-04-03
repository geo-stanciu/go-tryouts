package main

import (
	"fmt"
	"sync"
	"time"
)

// SourceLastRSS - LastRSS for source
type SourceLastRSS struct {
	SourceID    int       `sql:"rss_source_id"`
	LastRssDate time.Time `sql:"last_rss_date"`
	RssDate     []time.Time
}

// LastRssItems - Last Rss dates
type LastRssItems struct {
	sync.RWMutex
	RSS []*SourceLastRSS
}

// RssExists - Check if Rss exists in list
func (r *LastRssItems) RssExists(sourceID int) bool {
	for _, elem := range r.RSS {
		if elem.SourceID == sourceID {
			return true
		}
	}

	return false
}

// AddRSS - Add RSS elem
func (r *LastRssItems) AddRSS(s *SourceLastRSS) error {
	if r.RssExists(s.SourceID) {
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

		for _, dt := range elem.RssDate {
			if minLastRSS.IsZero() || minLastRSS.After(dt) {
				minLastRSS = dt
			}
		}

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

	tx.Commit()

	return nil
}
