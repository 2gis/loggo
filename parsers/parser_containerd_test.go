package parsers

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/2gis/loggo/common"
)

var errTest = errors.New("test")

var testCaseParserContainerD = []struct {
	name  string
	input string

	errExpected      error
	entryMapExpected common.EntryMap
}{
	{
		name:             "String does not contain 4 parts",
		input:            "{\"log\":\"hello world\"}",
		entryMapExpected: common.EntryMap(nil),
		errExpected:      errTest,
	},
	{
		name:  "Positive scenario log field map",
		input: "2020-09-10T07:00:03.585507743Z stdout F {\"hello\":\"world\",\"a\": 1,\"b\": null}",
		entryMapExpected: common.EntryMap{
			"hello":  "world",
			"a":      float64(1),
			"b":      nil,
			"stream": "stdout",
			"time":   "2020-09-10T07:00:03.585507743Z",
		},
	},
	{
		name:  "Positive scenario, log field plain string",
		input: "2020-09-10T07:00:03.585507743Z stdout F my message",
		entryMapExpected: common.EntryMap{
			"log":    "my message",
			"stream": "stdout",
			"time":   "2020-09-10T07:00:03.585507743Z",
		},
	},
}

func TestParseContainerDFormat(t *testing.T) {
	for _, testCase := range testCaseParserContainerD {

		t.Run(testCase.name, func(t *testing.T) {
			out, err := ParseContainerDFormat([]byte(testCase.input))
			assert.Equal(t, testCase.entryMapExpected, out)

			if testCase.errExpected != nil {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}
