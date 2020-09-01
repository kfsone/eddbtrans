package main

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/kfsone/eddbtrans"
)

func ErrIsBad(err error) {
	if err != nil {
		panic(err)
	}
}

func Must(err error) {
	ErrIsBad(err)
}

func main() {
	if len(os.Args) != 2 || os.Args[1] == "--help" || os.Args[1] == "-?" {
		log.Fatalf("Usage: %s <path>", os.Args[0])
	}
	path := os.Args[1]

	eddbtrans.SystemRegistry = eddbtrans.OpenDayCare()
	eddbtrans.FacilityRegistry = eddbtrans.OpenDayCare()

	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		convertSystems(path)
	}()
	go func() {
		defer wg.Done()
		convertStations(path)
	}()
	go func() {
		defer wg.Done()
		// Since the commodity list is tiny, lets just convert it first.
		convertCommodities(path)
		// This way we don't need a Daycare for commodity IDs.
		convertListings(path)
	}()

	wg.Wait()

	log.Printf("Finished entire conversion in %s.\n", time.Since(start))

	dc := eddbtrans.SystemRegistry
	log.Printf("Daycare stats: registered %d, queried %d, approved %d, queued %d, duplicated %d, denied %d\n",
		dc.Registered, dc.Queried, dc.Approved, dc.Queued, dc.Duplicate, dc.Queried-dc.Approved)
	if len(dc.Registry()) > 0 {
		stations := 0
		for _, items := range dc.Registry() {
			stations += len(items)
		}
		log.Printf("%d unrecognized systems referenced by %d stations.", len(dc.Registry()), stations)
	}

	eddbtrans.SystemRegistry.Close()
}
