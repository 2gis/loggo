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

func TestParseContainerDNginxTransforms(t *testing.T) {
	const prefix = "2020-09-10T07:00:03.585507743Z stdout F "

	testCases := []struct {
		name     string
		input    string
		expected common.EntryMap
	}{
		{
			name:     "field contains hypen",
			input:    `{"upstream_response_time":"-"}`,
			expected: common.EntryMap{LogKeyUpstreamResponseTime: "-"},
		},
		{
			name:  "single zero value given as string",
			input: `{"upstream_response_time":"0"}`,
			expected: common.EntryMap{
				LogKeyUpstreamResponseTime:            "0",
				LogKeyUpstreamResponseTimeReplacement: float64(0),
				LogKeyUpstreamResponseTimeTotal:       float64(0),
			},
		},
		{
			name:  "comma separated multiple value with spaces",
			input: `{"upstream_response_time":"0.009, 1.142, 1.222"}`,
			expected: common.EntryMap{
				LogKeyUpstreamResponseTime:            "0.009, 1.142, 1.222",
				LogKeyUpstreamResponseTimeReplacement: float64(1.222),
				LogKeyUpstreamResponseTimeTotal:       float64(2.373),
			},
		},
		{
			name:  "single float value",
			input: `{"upstream_response_time":1.222}`,
			expected: common.EntryMap{
				LogKeyUpstreamResponseTime:            float64(1.222),
				LogKeyUpstreamResponseTimeReplacement: float64(1.222),
				LogKeyUpstreamResponseTimeTotal:       float64(1.222),
			},
		},
		{
			name:     "value is given as json list, no transforms",
			input:    `{"upstream_response_time":[1.142, 1.222]}`,
			expected: common.EntryMap{LogKeyUpstreamResponseTime: []interface{}{1.142, 1.222}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			parser := CreateParserContainerDFormat(configFlattenTopLevel())
			out, err := parser([]byte(prefix + testCase.input))
			assert.NoError(t, err)

			expected := testCase.expected
			expected["stream"] = "stdout"
			expected["time"] = "2020-09-10T07:00:03.585507743Z"
			assert.Equal(t, expected, out)
		})
	}
}

func configFlattenTopLevel() configuration.ParserConfig {
	return configuration.ParserConfig{
		UserLogFieldsKey: "",
		CRIFieldsKey:     "",
		FlattenUserLog:   true,
		RawLogFieldKey:   "msg",
	}
}

func configFlattenSubDict() configuration.ParserConfig {
	return configuration.ParserConfig{
		UserLogFieldsKey: LogKeyLog,
		CRIFieldsKey:     "cri",
		FlattenUserLog:   true,
		RawLogFieldKey:   "msg",
	}
}
