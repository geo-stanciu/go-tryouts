package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	var xmlstr string
	var err error

	if len(os.Args) >= 2 {
		xmlstr, err = readFromFile(os.Args[1])
	} else {
		xmlstr, err = readFromURL("http://bnro.ro/nbrfxrates.xml")
	}

	fmt.Println(xmlstr)

	if err != nil {
		log.Fatal(err)
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
