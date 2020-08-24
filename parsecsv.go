package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
)

// ParseCsv will iterate over comma-separated lines in the source file,
// reordering the fields according to `fields`, and then sending the
// resulting slice of []byte to the returned channel.
//
// The first line of source must be the header fields.
//
// This is a very naive csv reader, it simply splits on commas.
//
func ParseCsv(source io.Reader, fields []string) (<-chan [][]byte, error) {
	// Read the first line and confirm it is a header field
	var fieldOrder = make([]int, 0, len(fields))
	channel := make(chan [][]byte, 16)

	scanner := bufio.NewScanner(source)
	if !scanner.Scan() {
		// Empty file
		close(channel)
		return channel, nil
	}

	headers := strings.Split(scanner.Text(), ",")
	for _, headerName := range headers {
		for dstIdx, fieldName := range fields {
			if strings.EqualFold(headerName, fieldName) {
				fieldOrder = append(fieldOrder, dstIdx)
				break
			}
		}
	}
	if len(fieldOrder) != len(fields) {
		close(channel)
		return nil, errors.New("missing fields in header")
	}

	go func() {
		scanner, fieldOrder, channel := scanner, fieldOrder, channel
		defer close(channel)
		for scanner.Scan() {
			result := make([][]byte, len(fields))
			columns := bytes.Split(scanner.Bytes(), []byte(","))
			for idx, value := range columns {
				result[fieldOrder[idx]] = value
			}
			channel <- result
		}
	}()

	return channel, nil
}
