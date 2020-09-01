package eddbtrans

import (
	"fmt"
	gom "github.com/kfsone/gomenacing/pkg/gomschema"
	"google.golang.org/protobuf/proto"
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
type parser func(io.Reader) (<-chan parsing.EntityPacket, error)

func ConvertFile(path, srcName, dstName string, hdrType gom.Header_Type, parserImpl parser, callback func(parsing.EntityPacket)) {
	srcPath, dstPath := filepath.Join(path, srcName), filepath.Join(path, dstName)

	// Open the input file
	srcFile, err := os.Open(srcPath)
	ErrIsBad(err)
	defer func() { ErrIsBad(srcFile.Close()) }()

	idxFilename := dstPath
	idxFile, err := os.Create(idxFilename)
	ErrIsBad(err)
	defer func() { ErrIsBad(idxFile.Close()) }()

	header := gom.Header{
		HeaderType: hdrType,
		Sizes:      make([]uint32, 0, 10240),
		Source:     "EDDB via " + srcPath,
		Userdata:   nil,
	}

	dataFilename := dstPath + ".dt"
	defer func() { ErrIsBad(os.Remove(dataFilename)) }()

	var start time.Time
	var messageCount, messageBytes int
	func() {
		dstFile, err := os.Create(dataFilename)
		ErrIsBad(err)
		defer func() { ErrIsBad(dstFile.Close()) }()

		start = time.Now()
		results, err := parserImpl(srcFile)
		ErrIsBad(err)

		for message := range results {
			header.Sizes = append(header.Sizes, uint32(len(message.Data)))

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
	}()

	data, err := proto.Marshal(&header)
	ErrIsBad(err)
	// We need to write the size of the header.
	_, err = idxFile.Write([]byte(fmt.Sprintf("GOMD%08x", len(data))))
	ErrIsBad(err)
	idxWritten, err := idxFile.Write(data)
	ErrIsBad(err)

	// Merge the .dt file into it.
	dataFile, err := os.Open(dataFilename)
	ErrIsBad(err)
	defer func() { ErrIsBad(dataFile.Close()) }()

	_, err = io.Copy(idxFile, dataFile)
	ErrIsBad(err)

	var avgSize = 0
	if messageCount > 0 {
		avgSize = messageBytes / messageCount
	}
	log.Printf("Converted %d %s to %s entries in %s, as %d bytes/%d avg. %d bytes idx %s.\n",
		messageCount, srcName, dataFilename, time.Since(start), messageBytes, avgSize, idxWritten, idxFilename)
}
