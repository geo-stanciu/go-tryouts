package main

import (
	"encoding/json"
	"os"
)

type rssSource struct {
	SourceName string `json:"SourceName"`
	Lang       string `json:"Lang"`
	Link       string `json:"Link"`
}

type rssSources struct {
	Rss []rssSource `json:"RSSSource"`
}

type configuration struct {
	DbType string `json:"DbType"`
	DbURL  string `json:"DbURL"`
	rssSources
}

func (c *configuration) ReadFromFile(cfgFile string) error {
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		return err
	}

	file, err := os.Open(cfgFile)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&c)
	if err != nil {
		return err
	}

	return nil
}
