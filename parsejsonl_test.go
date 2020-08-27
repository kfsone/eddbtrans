package eddbtrans

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"log"
	"strings"
	"testing"
	"time"
)

func captureLogOutput(callback func()) (logged [][]byte) {
	var buffer bytes.Buffer
	oldLog := log.Writer()
	defer log.SetOutput(oldLog)
	log.SetOutput(&buffer)

	callback()

	loggedText := buffer.Bytes()
	if len(loggedText) > 0 {
		logged = bytes.Split(bytes.TrimRight(loggedText, "\n"), []byte("\n"))
	}

	return logged
}

func Test_getJSONLines(t *testing.T) {
	// Passing an empty stream should just get us a closed channel.
	t.Run("Empty input", func(t *testing.T) {
		channel := getJSONLines(strings.NewReader(""))
		assert.NotNil(t, channel)
		_, ok := <-channel
		assert.False(t, ok)
	})

	// Should generate a log line, but only once.
	t.Run("Invalid input", func(t *testing.T) {
		var channel <-chan []byte
		var ok bool
		logged := captureLogOutput(func() {
			channel = getJSONLines(strings.NewReader("xyz\nabc\n"))
			// wait for the goroutine to finish
			_, ok = <-channel
		})
		assert.False(t, ok)
		if assert.Len(t, logged, 1) {
			assert.True(t, bytes.HasSuffix(logged[0], []byte("bad json: xyz")))
		}
	})

	t.Run("Valid input", func(t *testing.T) {
		var sources = []string{
			"[]",
			"{}",
			"1",
			"[2,    3, [], \"\", true,false]",
			"{\"a\":[1]  ,  \"b\": [2,3,{}]}",
		}
		var channel <-chan []byte
		var closed = false
		var lines = make([][]byte, 0, len(sources))
		var logged [][]byte
		go func() {
			logged = captureLogOutput(func() {
				text := strings.Join(sources, "\n") + "\n"
				channel = getJSONLines(strings.NewReader(text))
				for line := range channel {
					lines = append(lines, line)
				}
				closed = true
			})
		}()

		require.Eventually(t, func() bool { return closed }, time.Second, time.Microsecond*20)

		assert.Len(t, logged, 0)
		assert.Len(t, lines, len(sources))
		assert.Equal(t, strings.Join(sources, "\n"), string(bytes.Join(lines, []byte("\n"))))
	})
}

func tryParseJSONLines(input [][]byte, fields []string) (results [][]*gjson.Result, logged [][]byte) {
	lines := make(chan []byte, len(input) + 1)
	for i := range input {
		lines <- input[i]
	}
	close(lines)

	results = make([][]*gjson.Result, 0, len(input))
	var channel <-chan []*gjson.Result
	logged = captureLogOutput(func() {
		channel = parseJSONLines(lines, fields)
		for result := range channel {
			results = append(results, result)
		}
	})

	return
}

func Test_parseJSONLines(t *testing.T) {
	t.Run("Empty input", func(t *testing.T) {
		results, logged := tryParseJSONLines([][]byte{}, []string{})
		assert.Len(t, results, 0)
		assert.Len(t, logged, 0)
	})

	t.Run("Empty fields", func(t *testing.T) {
		results, logged := tryParseJSONLines([][]byte{[]byte("[1]")}, []string{})
		assert.Len(t, results, 0)
		assert.Len(t, logged, 1)
		assert.True(t, bytes.HasSuffix(logged[0], []byte("malformed jsonl: [1]")))
	})

	t.Run("Data", func(t *testing.T) {
		inputs := [][]byte{
			[]byte("{\"a\":1, \"b\":2, \"c\":[3,4,5]}"),
			[]byte("[2]"),
			[]byte("{\"a\":2}"),
			[]byte("{\"c\":47,\"a\":320}"),
		}
		results, logged := tryParseJSONLines(inputs, []string{"a","c"})
		assert.Len(t, results, 2)
		if assert.Len(t, logged, 2) {
			assert.True(t, bytes.HasSuffix(logged[0], []byte("malformed jsonl: [2]")))
			assert.True(t, bytes.HasSuffix(logged[1], []byte("missing \"c\" field in line: {\"a\":2}")))
		}
		if assert.Len(t, results[0], 2) {
			if assert.NotNil(t, results[0][0]) {
				assert.EqualValues(t, 1, results[0][0].Uint())
			}
			if assert.NotNil(t, results[0][1]) {
				assert.True(t, results[0][1].IsArray())
			}
		}
		if assert.Len(t, results[1], 2) {
			if assert.NotNil(t, results[1][0]) {
				assert.EqualValues(t, 320, results[1][0].Uint())
			}
			if assert.NotNil(t, results[1][1]) {
				assert.EqualValues(t, 47, results[1][1].Uint())
			}
		}
	})
}

func TestParseJSONLines(t *testing.T) {
	input := strings.Join([]string{
		`         {   "second":    "2nd","nth":"n","first"      : 3}`,
		`badnews`,
		`[1]`,
		`{"badnews":0}`,
		`{"first":[["a"],3.141,false], "second":2}`,
	}, "\n")
	var lines = make([][]*gjson.Result, 0, 8)
	logged := captureLogOutput(func() {
		channel := ParseJSONLines(strings.NewReader(input), []string{"first", "second"})
		for line := range channel {
			lines = append(lines, line)
		}
	})
	if 	assert.Len(t, logged, 3) {
		assert.True(t, bytes.HasSuffix(logged[0], []byte("bad json: badnews")))
		assert.True(t, bytes.HasSuffix(logged[1], []byte("malformed jsonl: [1]")))
		assert.True(t, bytes.HasSuffix(logged[2], []byte("missing \"first\" field in line: {\"badnews\":0}")))
	}
	if assert.Len(t, lines, 2) {
		if assert.Len(t, lines[0], 2) {
			assert.Equal(t, uint64(3), lines[0][0].Uint())
			assert.Equal(t, "2nd", lines[0][1].String())
		}
		if assert.Len(t, lines[1], 2) {
			if assert.True(t, lines[1][0].IsArray()) {
				assert.Len(t, lines[1][0].Array(), 3)
				assert.True(t, lines[1][0].Array()[0].IsArray())
				assert.Equal(t, 3.141, lines[1][0].Array()[1].Float())
				assert.Equal(t, false, lines[1][0].Array()[2].Bool())
			}
			assert.Equal(t, uint64(2), lines[1][1].Uint())
		}
	}
}