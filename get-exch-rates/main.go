package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Rate struct {
	Currency   string `xml:"currency,attr"`
	Multiplier string `xml:"multiplier,attr"`
	Rate       string `xml:",chardata"`
}

type Cube struct {
	Date string `xml:"date,attr"`
	Rate []Rate
}

type Header struct {
	Publisher      string `xml:"Publisher"`
	PublishingDate string `xml:"PublishingDate"`
	MessageType    string `xml:"MessageType"`
}

type Body struct {
	Subject      string `xml:"Subject"`
	OrigCurrency string `xml:"OrigCurrency"`
	Cube         []Cube
}

type Query struct {
	XMLName xml.Name `xml:"DataSet"`
	Header  Header   `xml:"Header"`
	Body    Body     `xml:"Body"`
}

func main() {
	var xmlBytes []byte
	var err error

	if len(os.Args) >= 2 {
		xmlBytes, err = readBytesFromFile(os.Args[1])
	} else {
		xmlBytes, err = readBytesFromURL("http://bnro.ro/nbrfxrates.xml")
	}

	if err != nil {
		log.Fatal(err)
	}

	var q Query

	err = xml.Unmarshal(xmlBytes, &q)

	if err != nil {
		log.Fatal(err)
	}

	for _, cube := range q.Body.Cube {
		fmt.Printf("Date: %v\n", cube.Date)

		for _, rate := range cube.Rate {
			multiplier := 1.0
			exchRate := 1.0

			if len(rate.Multiplier) > 0 {
				multiplier, err = strconv.ParseFloat(rate.Multiplier, 64)

				if err != nil {
					log.Fatal(err)
				}
			}

			exchRate, err = strconv.ParseFloat(rate.Rate, 64)

			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("%v: %.6f\n", rate.Currency, exchRate/multiplier)
		}
	}
}

func readBytesFromURL(url string) ([]byte, error) {
	response, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	buf, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	return buf, nil
}

func readBytesFromFile(filename string) ([]byte, error) {
	f, err := os.Open(filename)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	buf, err := ioutil.ReadAll(f)

	if err != nil {
		return nil, err
	}

	return buf, nil
}
