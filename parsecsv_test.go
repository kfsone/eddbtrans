package eddbtrans

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func isChannelClosed(channel <-chan [][]byte) bool {
	select {
	case _, ok := <-channel:
		return !ok
	default:
		return false
	}
}

func Test_getFieldOrder(t *testing.T) {
	t.Run("Empty v Empty", func(t *testing.T) {
		fieldOrder, fields := getFieldOrder([]string{}, "")
		if assert.NotNil(t, fieldOrder) {
			assert.Len(t, fieldOrder, 0)
			assert.Equal(t, 0, fields)
		}
	})

	t.Run("Headings without fields", func(t *testing.T) {
		fieldOrder, fields := getFieldOrder([]string{}, "first,b,second,3")
		if assert.NotNil(t, fieldOrder) {
			assert.Len(t, fieldOrder, 0)
			assert.Equal(t, 4, fields)
		}
	})

	t.Run("Fields without headings", func(t *testing.T) {
		fieldOrder, fields := getFieldOrder([]string{"first", "second"}, "")
		if assert.NotNil(t, fieldOrder) {
			assert.Len(t, fieldOrder, 0)
			assert.Equal(t, 0, fields)
		}
	})

	t.Run("Undermatch", func(t *testing.T) {
		fieldOrder, fields := getFieldOrder([]string{"second"}, "first,second,third")
		if assert.NotNil(t, fieldOrder) {
			assert.Len(t, fieldOrder, 1)
			assert.Equal(t, 1, fieldOrder[0])
			assert.Equal(t, 3, fields)
		}
	})

	t.Run("Matches", func(t *testing.T) {
		fieldOrder, fields := getFieldOrder([]string{"second", "third"}, "first,third,fifth,second,4")
		if assert.NotNil(t, fieldOrder) {
			assert.Equal(t, []int{3, 1}, fieldOrder)
			assert.Equal(t, 5, fields)
		}
	})
}

func TestParseCsv(t *testing.T) {
	t.Run("Empty stream", func(t *testing.T) {
		// Passing an empty stream should produce a closed channel with no error.
		reader := strings.NewReader("")
		channel, err := ParseCsv(reader, []string{})
		if assert.Nil(t, channel) {
			assert.Equal(t, io.EOF, err)
		}
	})

	t.Run("No matching headers", func(t *testing.T) {
		reader := strings.NewReader("x,y,z")
		channel, err := ParseCsv(reader, []string{"a"})
		assert.Error(t, err)
		assert.Nil(t, channel)
	})

	t.Run("Missing headers", func(t *testing.T) {
		reader := strings.NewReader("a,x,y,z")
		channel, err := ParseCsv(reader, []string{"a", "b", "z"})
		assert.Error(t, err)
		assert.Nil(t, channel)
	})

	t.Run("Empty with headers", func(t *testing.T) {
		reader := strings.NewReader("a,b,c,")
		channel, err := ParseCsv(reader, []string{"a", "c"})
		assert.Nil(t, err)
		assert.NotNil(t, channel)
		assert.Eventually(t, func() bool { return isChannelClosed(channel) }, time.Millisecond*10, time.Microsecond*20)
	})

	t.Run("Noise with headers", func(t *testing.T) {
		reader := strings.NewReader("a,b,c,\n\n\n\n\n\n\n\n\ninvalid\n\n\n\n\n")
		channel, err := ParseCsv(reader, []string{"a", "c"})
		assert.Nil(t, err)
		assert.NotNil(t, channel)
		assert.Eventually(t, func() bool { return isChannelClosed(channel) }, time.Millisecond*10, time.Microsecond*20)
	})

	t.Run("Nominal use", func(t *testing.T) {
		reader := strings.NewReader("third,first,fourth,second\n\"this\",1,\"is\",\"number two\"")
		channel, err := ParseCsv(reader, []string{"first", "second"})
		assert.Nil(t, err)
		assert.NotNil(t, channel)

		var result [][]byte = nil
		go func() {
			result = <-channel
		}()

		assert.Eventually(t, func() bool { return result != nil }, time.Millisecond*100, time.Microsecond*50)
		if assert.Len(t, result, 2) {
			assert.Equal(t, []byte("1"), result[0])
			assert.Equal(t, []byte("number two"), result[1])
		}

		assert.Eventually(t, func() bool { return isChannelClosed(channel) }, time.Millisecond*100, time.Millisecond*50)
	})

	t.Run("Lines", func(t *testing.T) {
		reader := strings.NewReader("3rd,1st,4th,2nd\n\n-,1,-,\"2\",\n\n\ninvalid,invalid\n-,\"first,-,second\"\n\n")
		channel, err := ParseCsv(reader, []string{"1st", "2nd"})
		require.Nil(t, err)
		assert.NotNil(t, channel)

		var result [][]byte = nil
		go func() {
			result = <-channel
		}()

		assert.Eventually(t, func() bool { return result != nil }, time.Millisecond*50, time.Microsecond*20)
		if assert.Len(t, result, 2) {
			assert.Equal(t, []byte("1"), result[0])
			assert.Equal(t, []byte("2"), result[1])
		}

		result = nil
		go func() {
			result = <-channel
		}()

		assert.Eventually(t, func() bool { return result != nil }, time.Millisecond*50, time.Microsecond*20)
		if assert.Len(t, result, 2) {
			assert.Equal(t, []byte("first"), result[0])
			assert.Equal(t, []byte("second"), result[1])
		}

		assert.Eventually(t, func() bool { return isChannelClosed(channel) }, time.Millisecond*50, time.Millisecond*20)
	})
}
