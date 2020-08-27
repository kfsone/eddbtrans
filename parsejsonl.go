package eddbtrans

import (
	"bufio"
	"io"
	"log"

	"github.com/tidwall/gjson"
)

func getJSONLines(source io.Reader) <-chan []byte {
	channel := make(chan []byte, 1)
	go func() {
		defer close(channel)

		badLines := false
		scanner := bufio.NewScanner(source)

		for scanner.Scan() {
			line := scanner.Bytes()
			if !gjson.ValidBytes(line) {
				if !badLines {
					log.Printf("bad json: %s", string(line))
					badLines = true
				}
				continue
			}
			channel <- line
		}
	}()

	return channel
}

func parseJSONLines(lines <-chan []byte, fields []string) <-chan []*gjson.Result {
	channel := make(chan []*gjson.Result, 1)
	go func() {
		defer close(channel)
		badLines := false
		for line := range lines {
			result := gjson.ParseBytes(line)
			if !result.IsObject() {
				if !badLines {
					log.Printf("malformed jsonl: %s", string(line))
					badLines = true
				}
				continue
			}
			jsonFields := result.Map()
			slice := make([]*gjson.Result, len(fields))
			invalid := false
			for idx, field := range fields {
				if jsonField, ok := jsonFields[field]; ok {
					slice[idx] = &jsonField
				} else {
					if !invalid {
						log.Printf("missing \"%s\" field in line: %s", field, line)
						invalid = true
					}
					break
				}
			}
			if !invalid {
				channel <- slice
			}
		}
	}()

	return channel
}

// ParseJSONLines will read lines from a `.jsonl` file, json parse them and then
// send a slice of []byte arrays to the channel representing the columns
// requested in `fields`.
//
func ParseJSONLines(source io.Reader, fields []string) <-chan []*gjson.Result {
	return parseJSONLines(getJSONLines(source), fields)
}
