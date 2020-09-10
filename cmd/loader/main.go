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
	gomFile, err := gom.OpenGOMFile(file)
	if err != nil {
		panic(err)
	}
	defer gomFile.Close()
	items, err := gomFile.Load()
	if err != nil {
		panic(err)
	}
	log.Printf("Took %s\n", time.Since(start))

	for idx, item := range items {
		if idx >= 3 {
			fmt.Println("...")
			return
		}
		fmt.Printf("%v\n", item)
	}
}
