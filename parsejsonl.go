package eddbtrans

import (
	"bufio"
	"io"
	"log"

	"github.com/tidwall/gjson"
)

type JsonLinesChannel chan []byte
type JsonResultChannel chan []*gjson.Result

func getJsonLines(source io.Reader) JsonLinesChannel {
	channel := make(JsonLinesChannel, 8)
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

func parseJsonLines(lines JsonLinesChannel, fields []string) JsonResultChannel {
	channel := make(JsonResultChannel, 8)
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
			for idx, field := range fields {
				jsonField := jsonFields[field]
				slice[idx] = &jsonField
			}
			channel <- slice
		}
	}()

	return channel
}

// ParseJsonLines will read lines from a `.jsonl` file, json parse them and then
// send a slice of []byte arrays to the channel representing the columns
// requested in `fields`.
//
func ParseJsonLines(source io.Reader, fields []string) JsonResultChannel {
	return parseJsonLines(getJsonLines(source), fields)
}
