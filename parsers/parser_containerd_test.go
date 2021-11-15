package parsers

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/configuration"
)

type testCaseParserContainerD struct {
	name  string
	input string

	config configuration.ParserConfig

	errExpected      error
	entryMapExpected common.EntryMap
}

var errTest = errors.New("test")

var testCasesParserContainerD = []testCaseParserContainerD{
	{
		name:             "String does not contain 4 parts",
		input:            "{\"log\":\"hello world\"}",
		entryMapExpected: common.EntryMap(nil),
		errExpected:      errTest,
	},
	{
		name:   "Positive scenario log field map",
		config: configFlattenTopLevel(),
		input:  "2020-09-10T07:00:03.585507743Z stdout F {\"hello\":\"world\",\"a\": 1,\"b\": null}",
		entryMapExpected: common.EntryMap{
			"hello":  "world",
			"a":      float64(1),
			"b":      nil,
			"stream": "stdout",
			"time":   "2020-09-10T07:00:03.585507743Z",
		},
	},
	{
		name:   "Positive scenario, log field plain string",
		config: configFlattenTopLevel(),
		input:  "2020-09-10T07:00:03.585507743Z stdout F my message",
		entryMapExpected: common.EntryMap{
			"msg":    "my message",
			"stream": "stdout",
			"time":   "2020-09-10T07:00:03.585507743Z",
		},
	},
	{
		name:   "Positive scenario, log field plain string",
		config: configFlattenSubDict(),
		input:  "2020-09-10T07:00:03.585507743Z stdout F {\"hello\":\"world\",\"a\": 1,\"b\": null}",
		entryMapExpected: common.EntryMap{
			"log": common.EntryMap{
				"hello": "world",
				"a":     float64(1),
				"b":     nil,
			},
			"cri": common.EntryMap{
				"stream": "stdout",
				"time":   "2020-09-10T07:00:03.585507743Z",
			},
		},
	},
}

func TestParseContainerDFormat(t *testing.T) {
	for _, testCase := range testCasesParserContainerD {

		t.Run(testCase.name, func(t *testing.T) {
			parser := CreateParserContainerDFormat(testCase.config)
			out, err := parser([]byte(testCase.input))
			assert.Equal(t, testCase.entryMapExpected, out)

			if testCase.errExpected != nil {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}
