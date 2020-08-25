package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kfsone/eddbtrans"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] == "--help" || os.Args[1] == "-?" {
		log.Fatalf("Usage: %s <filename>", os.Args[0])
	}
	filename := os.Args[1]
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	///TODO: Pass a context for cancellation.
	results, err := eddbtrans.ParseSystemsPopulatedJsonl(file)
	if err != nil {
		panic(err)
	}
	for message := range results {
		fmt.Printf("%d: %d bytes\n", message.ObjectId, len(message.Data))
		break
	}
}
