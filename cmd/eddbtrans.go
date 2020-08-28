package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/kfsone/eddbtrans"
)

func parseCommodity(path string) {
	filename := filepath.Join(path, "commodities.json")
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	start := time.Now()
	results, err := eddbtrans.ParseCommodityJson(file)
	if err != nil {
		panic(err)
	}
	messageCount := 0
	for message := range results {
		messageCount++
		if messageCount <= 1 {
			fmt.Printf("Commodity #%d: %d: %d bytes.\n", messageCount, message.ObjectId, len(message.Data))
		}
	}
	fmt.Printf("Converted %d commodities in %s.\n", messageCount, time.Since(start))
}

func parseSystems(path string) {
	filename := filepath.Join(path, "systems_populated.jsonl")
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	start := time.Now()
	results, err := eddbtrans.ParseSystemsPopulatedJSONL(file)
	if err != nil {
		panic(err)
	}
	messageCount := 0
	for message := range results {
		messageCount++
		if messageCount <= 1 {
			fmt.Printf("System #%d: %d: %d bytes.\n", messageCount, message.ObjectId, len(message.Data))
		}
	}
	fmt.Printf("Converted %d systems in %s.\n", messageCount, time.Since(start))
}

func parseStations(path string) {
	filename := filepath.Join(path, "stations.jsonl")
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	start := time.Now()
	results, err := eddbtrans.ParseStationJSONL(file)
	if err != nil {
		panic(err)
	}
	messageCount := 0
	for message := range results {
		messageCount++
		if messageCount <= 1 {
			fmt.Printf("Station #%d: %d: %d bytes.\n", messageCount, message.ObjectId, len(message.Data))
		}
	}
	fmt.Printf("Converted %d stations in %s.\n", messageCount, time.Since(start))
}

func main() {
	if len(os.Args) != 2 || os.Args[1] == "--help" || os.Args[1] == "-?" {
		log.Fatalf("Usage: %s <path>", os.Args[0])
	}
	path := os.Args[1]
	start := time.Now()

	// We'll run all three conversions simultaneously in the background, so we need a way
	// to wait for them to complete. This is the go "WaitGroup".
	var wg sync.WaitGroup
	// Add 3 processes expected.
	wg.Add(3)
	go func () {
		defer wg.Done()
		parseCommodity(path)
	}()
	go func () {
		defer wg.Done()
		parseSystems(path)
	}()
	go func () {
		defer wg.Done()
		parseStations(path)
	}()

	wg.Wait()
	fmt.Printf("Finished entire conversion in %s.\n", time.Since(start))
}
