package main

import (
	"bufio"
	"bytes"
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
	var xmlstr string
	var err error

	if len(os.Args) >= 2 {
		xmlstr, err = readFromFile(os.Args[1])
	} else {
		xmlstr, err = readFromURL("http://bnro.ro/nbrfxrates.xml")
	}

	if err != nil {
		log.Fatal(err)
	}

	var q Query

	err = xml.Unmarshal([]byte(xmlstr), &q)

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

			fmt.Printf("%v: %.4f\n", rate.Currency, exchRate/multiplier)
		}
	}
}

func readFromURL(url string) (string, error) {
	response, err := http.Get(url)

	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	buf, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return "", err
	}

	xmlstr := fmt.Sprintf("%s", buf)

	return xmlstr, nil
}

func readFromFile(filename string) (string, error) {
	f, err := os.Open(filename)

	if err != nil {
		return "", err
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	var buffer bytes.Buffer

	for scanner.Scan() {
		buffer.WriteString(scanner.Text())
		buffer.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return buffer.String(), nil
}
