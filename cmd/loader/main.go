package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	gom "github.com/kfsone/gomenacing/pkg/gomschema"
)

var filename = flag.String("file", "", "Path/file of the .gom file to load.")

func main() {
	flag.Parse()
	if *filename == "" {
		panic("No -file specified.")
	}
	file, err := os.Open(*filename)
	if err != nil {
		panic(err)
	}

	start := time.Now()
	items, err := gom.ReadGOMFile(file)
	if err != nil {
		panic(err)
	}
	log.Printf("Took %s\n", time.Since(start))

	switch list := items.(type) {
	case []gom.Commodity:
		for idx, item := range list {
			if idx >= 10 {
				fmt.Println("...")
				return
			}
			fmt.Printf("%v\n", item)
		}
	case []gom.System:
		for idx, item := range list {
			if idx >= 10 {
				fmt.Println("...")
				return
			}
			fmt.Printf("%v\n", item)
		}
	case []gom.Facility:
		for idx, item := range list {
			if idx >= 10 {
				fmt.Println("...")
				return
			}
			fmt.Printf("%v\n", item)
		}
	case []gom.FacilityListing:
		for idx, item := range list {
			if idx >= 10 {
				fmt.Println("...")
				return
			}
			fmt.Printf("%v\n", item)
		}
	default:
		fmt.Printf("Unrecognized type.")
	}
}
