package main

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/geo-stanciu/go-utils/utils"
)

// RssLink - statistics for each rss link
type RssLink struct {
	Link     string
	RssDate  time.Time
	NewItems int
}

// RSSFeed - Last RSS items for source
type RSSFeed struct {
	sync.RWMutex
	SourceID          int `sql:"rss_source_id"`
	LoweredSourceName string
	LastUpdate        time.Time `sql:"last_rss_date"`
	Links             []*RssLink
	rssHash           map[string]bool
}

// Initialize - Initialize
func (s *RSSFeed) Initialize(sourceName string) {
	epochStart, _ := utils.String2date("1970-01-01", utils.UTCDate)
	s.SourceID = -1
	s.LoweredSourceName = strings.ToLower(sourceName)
	s.LastUpdate = epochStart
	s.rssHash = make(map[string]bool)
}

// GetLink - Get Rss link statistics by link
func (s *RSSFeed) GetLink(lnk string) *RssLink {
	for _, elem := range s.Links {
		if elem.Link == lnk {
			return elem
		}
	}

	return nil
}

// RssExists - check if the hash of title##link was already added in current session
func (s *RSSFeed) RssExists(title string, lnk string) bool {
	key := fmt.Sprintf("%s##%s", title, lnk)

	h := sha256.New()
	h.Write([]byte(key))

	hash := fmt.Sprintf("%x", h.Sum(nil))

	if _, ok := s.rssHash[hash]; ok {
		return true
	}

	s.rssHash[hash] = true

	return false
}
