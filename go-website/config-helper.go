package main

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	Port     string `json:"Port"`
	DbType   string `json:"DbType"`
	DbURL    string `json:"DbURL"`
	Timezone string `json:"Timezone"`
	IsHTTPS  bool   `json:"isHTTPS"`
}

func (c *Configuration) ReadFromFile(cfgFile string) error {
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
