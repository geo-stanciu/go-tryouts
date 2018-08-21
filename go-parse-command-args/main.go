package main

import (
	"flag"
	"fmt"
)

func main() {
	cfgPtr := flag.String("c", "conf.json", "config file")

	flag.Parse()

	fmt.Println("c:", *cfgPtr)
}
