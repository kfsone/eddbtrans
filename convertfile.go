package eddbtrans

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/kfsone/gomenacing/pkg/parsing"
)

// ErrIsBad will panic if error is not nil.
func ErrIsBad(err error) {
	if err != nil {
		panic(err)
	}
}

// parser is an alias for functions taking a reader and returning a stream of entity packets.
type parser func(io.Reader) (<- chan parsing.EntityPacket, error)

func ConvertFile(path, srcName, dstName string, parserImpl parser, callback func(parsing.EntityPacket)) {
	srcPath, dstPath := filepath.Join(path, srcName), filepath.Join(path, dstName)

	// Open the input file
	srcFile, err := os.Open(srcPath)
	ErrIsBad(err)
	defer func() { ErrIsBad(srcFile.Close()) }()

	dstFile, err := os.Create(dstPath)
	ErrIsBad(err)
	defer func() { ErrIsBad(dstFile.Close()) }()

	start := time.Now()
	results, err := parserImpl(srcFile)
	ErrIsBad(err)

	var messageCount, messageBytes int
	for message := range results {
		// Save the packet, barf if it fails on us.
		written, err := dstFile.Write(message.Data)
		if err == nil && written != len(message.Data) {
			err = io.ErrUnexpectedEOF
		}
		ErrIsBad(err)

		messageCount++
		messageBytes += len(message.Data)

		// Call the callback if there is one.
		if callback != nil {
			callback(message)
		}
	}

	var avgSize = 0
	if messageCount > 0 {
		avgSize = messageBytes / messageCount
	}
	log.Printf("Converted %d %s to %s entries in %s, as %d bytes/%d avg.\n",
				messageCount, srcName, dstName, time.Since(start), messageBytes, avgSize)
}
