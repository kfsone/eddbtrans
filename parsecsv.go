package eddbtrans

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
)

// getFieldOrder will identify the order of the comma-separated
// heading values in `line` as required to represent `fields`,
// as well as the number of headings in `line`.
func getFieldOrder(fields []string, line string) ([]int, int) {
	var fieldOrder = make([]int, 0, len(fields))
	if len(line) == 0 {
		return fieldOrder, 0
	}
	headers := strings.Split(line, ",")
	for _, fieldName := range fields {
		for headerNo, headerName := range headers {
			if strings.EqualFold(fieldName, headerName) {
				fieldOrder = append(fieldOrder, headerNo)
				break
			}
		}
	}

	return fieldOrder, len(headers)
}

// ParseCsv will iterate over comma-separated lines in the source file,
// reordering the fields according to `fields`, and then sending the
// resulting slice of []byte to the returned channel.
//
// The first line of source must be the header fields.
//
// This is a very naive csv reader, it simply splits on commas.
//
func ParseCsv(source io.Reader, fields []string) (<-chan [][]byte, error) {
	// Scan for a first line, which should contain the headers.
	scanner := bufio.NewScanner(source)
	if !scanner.Scan() {
		// If we couldn't scan one line, the file is empty. Close the channel
		// and return it with no error.
		return nil, io.EOF
	}

	// Map the csv column order to the requested fields order.
	fieldOrder, numHeadings := getFieldOrder(fields, scanner.Text())
	if len(fieldOrder) != len(fields) {
		return nil, errors.New("missing fields in header")
	}

	channel := make(chan [][]byte, 16)
	go func() {
		// Break remaining lines up by
		scanner, fieldOrder, channel := scanner, fieldOrder, channel
		var separator = []byte(",")
		defer close(channel)
		for scanner.Scan() {
			result := make([][]byte, len(fields))
			columns := bytes.Split(scanner.Bytes(), separator)
			if len(columns) < numHeadings {
				continue
			}
			for fieldNo, columnIdx := range fieldOrder {
				value := columns[columnIdx]
				if len(value) > 0 && value[0] == '"' {
					value = value[1:]
				}
				if len(value) > 0 && value[len(value)-1] == '"' {
					value = value[:len(value)-1]
				}
				result[fieldNo] = value
			}
			channel <- result
		}
	}()

	return channel, nil
}
