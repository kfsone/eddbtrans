package main

import (
	"bufio"
	"github.com/tidwall/gjson"
	"io"
	"log"
)

type JsonLinesChannel chan []byte
type JsonResultChannel chan []*gjson.Result

func getJsonLines(source io.Reader, channel JsonLinesChannel) {
	defer close(channel)

	badLines := 0
	scanner := bufio.NewScanner(source)

	for scanner.Scan() {
		line := scanner.Bytes()
		if !gjson.ValidBytes(line) {
			if badLines == 0 {
				log.Printf("bad json: %s", string(line))
			}
			badLines += 1
			continue
		}
		channel <- line
	}
}

func parseJsonLines(lines JsonLinesChannel, channel JsonResultChannel, fields []string) {
	defer close(channel)
	badLines := 0

	for line := range lines {
		result := gjson.ParseBytes(line)
		if !result.IsObject() {
			if badLines == 0 {
				log.Printf("malformed jsonl: %s", string(line))
			}
			badLines += 1
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
}

// ParseJsonLines will read lines from a `.jsonl` file, json parse them and then
// send a slice of []byte arrays to the channel representing the columns
// requested in `fields`.
//
func ParseJsonLines(source io.Reader, fields []string) JsonResultChannel {
	lines := make(JsonLinesChannel, 8)
	go getJsonLines(source, lines)

	channel := make(JsonResultChannel, 8)
	go parseJsonLines(lines, channel, fields)

	return channel
}
