package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/geo-stanciu/go-utils/utils"
)

type textfile struct {
	Name string
	Body string
}

func main() {
	var buf bytes.Buffer

	var files = []textfile{
		{"readme.txt", "This archive contains some text files."},
		{"gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
		{"todo.txt", "Get animal handling licence.\nWrite more examples."},
	}

	// zip
	zw := utils.NewZipWriter(&buf)

	for _, file := range files {
		err := zw.AddEntry(file.Name, []byte(file.Body))
		if err != nil {
			log.Fatal(err)
		}
	}

	// call close prior trying to use the newly created archive
	zw.Close()

	// get byte slice
	content := buf.Bytes()
	buf.Reset()

	// unzip
	zr, err := utils.NewZipReader(content)
	if err != nil {
		log.Fatal(err)
	}
	defer zr.Close()

	for zr.HasNextEntry() {
		f, err := zr.GetNextEntry()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("\nContents of %s:\n", f)
		err = zr.ReadCurrentEntry(os.Stdout)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println()
	}
}
