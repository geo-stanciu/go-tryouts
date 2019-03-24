package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var (
	currentDir string
)

func init() {
	currentDir = filepath.Dir(os.Args[0])
}

func main() {
	cfgPtr := flag.String("c", fmt.Sprintf("%s/conf.json", currentDir), "config file")

	flag.Parse()

	fmt.Println("c:", *cfgPtr)
}
