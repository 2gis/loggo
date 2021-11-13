package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/configuration"
)

type testCaseParserDocker struct {
	name string

	config           configuration.ParserConfig
	input            string
	entryMapExpected common.EntryMap
	errExpected      error
}

var (
	testCasesParserDocker = []testCaseParserDocker{
		{
			name:             "Positive, flattening to top-level",
			input:            "{\"log\":\"hello world\", \"key1\":1}",
			entryMapExpected: common.EntryMap{"log": "hello world", "key1": float64(1)},
		},
		{
			name:             "log field is missing",
			input:            "{\"somekey\":\"hello world\"}",
			entryMapExpected: nil,
			errExpected:      errTest,
		},
		{
			name:             "invalid json syntax",
			input:            "{log\":\"hello world\"}",
			entryMapExpected: nil,
			errExpected:      errTest,
		},
		{
			name:             "inner map positive",
			input:            `{"log": "{\"hello\":\"world\",\"a\": 1,\"b\": null}"}`,
			entryMapExpected: common.EntryMap{"hello": "world", "a": float64(1), "b": nil},
		},
	}

	testCasesParserDockerNginxTransform = []testCaseParserDocker{
		{
			name:  "field contains hypen",
			input: `{"log": "{\"upstream_response_time\":\"-\"}"}`,
			entryMapExpected: common.EntryMap{
				LogKeyUpstreamResponseTime: "-",
			},
		},
		{
			name:  "single zero value given as string",
			input: `{"log": "{\"upstream_response_time\":\"0\"}"}`,
			entryMapExpected: common.EntryMap{
				LogKeyUpstreamResponseTime:            "0",
				LogKeyUpstreamResponseTimeReplacement: float64(0),
				LogKeyUpstreamResponseTimeTotal:       float64(0),
			},
		},
		{
			name:  "single non-zero value given as string",
			input: `{"log": "{\"upstream_response_time\":\"0.009\"}"}`,
			entryMapExpected: common.EntryMap{
				LogKeyUpstreamResponseTime:            "0.009",
				LogKeyUpstreamResponseTimeReplacement: float64(0.009),
				LogKeyUpstreamResponseTimeTotal:       float64(0.009),
			},
		},
		{
			name:  "comma separated multiple value without spaces",
			input: `{"log": "{\"upstream_response_time\":\"0.009,1.142\"}"}`,
			entryMapExpected: common.EntryMap{
				LogKeyUpstreamResponseTime:            "0.009,1.142",
				LogKeyUpstreamResponseTimeReplacement: float64(1.142),
				LogKeyUpstreamResponseTimeTotal:       float64(1.151),
			},
		},
		{
			name:  "comma separated multiple value with spaces",
			input: `{"log": "{\"upstream_response_time\":\"0.009, 1.142, 1.222\"}"}`,
			entryMapExpected: common.EntryMap{
				LogKeyUpstreamResponseTime:            "0.009, 1.142, 1.222",
				LogKeyUpstreamResponseTimeReplacement: float64(1.222),
				LogKeyUpstreamResponseTimeTotal:       float64(2.373),
			},
		},
		// not sure, maybe we should cast it to string anyway for uniformity? leaved as is to preserve early logic
		{
			name:  "single float value",
			input: `{"log": "{\"upstream_response_time\":1.222}"}`,
			entryMapExpected: common.EntryMap{
				LogKeyUpstreamResponseTime:            float64(1.222),
				LogKeyUpstreamResponseTimeReplacement: float64(1.222),
				LogKeyUpstreamResponseTimeTotal:       float64(1.222),
			},
		},
		{
			name:  "value is given as json list, passed as json list, no transforms",
			input: `{"log": "{\"upstream_response_time\":[1.142, 1.222]}"}`,
			entryMapExpected: common.EntryMap{
				LogKeyUpstreamResponseTime: []interface{}{1.142, 1.222},
			},
		},
		{
			name:  "upstream_response_time contains dict",
			input: `{"log": "{\"upstream_response_time\":{\"value\": 1.142}}"}`,
			entryMapExpected: common.EntryMap{
				LogKeyUpstreamResponseTime + ".value": float64(1.142),
			},
		},
	}

	testCasesParserDockerSubmaps = []testCaseParserDocker{
		{
			name: "user log set to separate field, not flattened",
			config: configuration.ParserConfig{
				UserLogFieldsKey: "user_log",
				CRIFieldsKey:     "docker",
				FlattenUserLog:   false,
			},
			input: `{"time": "2018-01-09T05:08:03.100481875Z", "log": "{\"time\":{\"value\": 1.142}}"}`,
			entryMapExpected: common.EntryMap{
				"user_log": common.EntryMap{"time": map[string]interface{}{"value": 1.142}},
				"docker":   common.EntryMap{"time": "2018-01-09T05:08:03.100481875Z"},
			},
		},
		{
			name: "user log set to separate field, flattened",
			config: configuration.ParserConfig{
				UserLogFieldsKey: "user_log",
				CRIFieldsKey:     "docker",
				FlattenUserLog:   true,
			},
			input: `{"time": "2018-01-09T05:08:03.100481875Z", "log": "{\"time\":{\"value\": 1.142}}"}`,
			entryMapExpected: common.EntryMap{
				"user_log": common.EntryMap{"time.value": 1.142},
				"docker":   common.EntryMap{"time": "2018-01-09T05:08:03.100481875Z"},
			},
		},
		{
			name: "flattening is on, user log field is empty, user log expected in top level dict",
			config: configuration.ParserConfig{
				UserLogFieldsKey: "",
				CRIFieldsKey:     "docker",
				FlattenUserLog:   true,
			},
			input: `{"time": "2018-01-09T05:08:03.100481875Z", "log": "{\"time\":{\"value\": 1.142}}"}`,
			entryMapExpected: common.EntryMap{
				"time.value": 1.142,
				"docker":     common.EntryMap{"time": "2018-01-09T05:08:03.100481875Z"},
			},
		},
		{
			name: "flattening is off, user log field is empty, user log expected in log field of top level dict",
			config: configuration.ParserConfig{
				UserLogFieldsKey: "",
				CRIFieldsKey:     "docker",
				FlattenUserLog:   false,
			},
			input: `{"time": "2018-01-09T05:08:03.100481875Z", "log": "{\"time\":{\"value\": 1.142}}"}`,
			entryMapExpected: common.EntryMap{
				"log":    common.EntryMap{"time": map[string]interface{}{"value": 1.142}},
				"docker": common.EntryMap{"time": "2018-01-09T05:08:03.100481875Z"},
			},
		},
		{
			name: "flattening is on, user log field key and docker fields key is empty, " +
				"user log expected to override docker variables (legacy behavior)",
			config: configuration.ParserConfig{
				UserLogFieldsKey: "",
				CRIFieldsKey:     "",
				FlattenUserLog:   true,
			},
			input: `{"time": "2018-01-09T05:08:03.100481875Z", "log": "{\"time\": 1.142, \"test\": \"value\"}"}`,
			entryMapExpected: common.EntryMap{
				"time": 1.142,
				"test": "value",
			},
		},
	}

	testCasesParserDockerUnpackingControl = []testCaseParserDocker{
		{
			name:             "Different keys to unpack",
			input:            `{"log": "{\"hello\":\"world\",\"a\": 1,\"b\": null}", "a": "else"}`,
			entryMapExpected: common.EntryMap{
				"log": common.EntryMap{"hello": "world", "a": float64(1), "b": nil},
				"cri": common.EntryMap{"a": "else"},
			},
			config: configFlattenSubDict(),
		},
	}
)

func TestParseLineBasic(t *testing.T) {
	for _, testCase := range testCasesParserDocker {
		t.Run(testCase.name, func(t *testing.T) {
			parser := CreateParserDockerFormat(configFlattenTopLevel())

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

func TestParseLineNginxTransforms(t *testing.T) {
	for _, testCase := range testCasesParserDockerNginxTransform {
		t.Run(testCase.name, func(t *testing.T) {
			parser := CreateParserDockerFormat(configFlattenTopLevel())

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

func TestSeparateDockerUserLogFieldsSet(t *testing.T) {
	for _, testCase := range testCasesParserDockerSubmaps {
		t.Run(testCase.name, func(t *testing.T) {
			parser := CreateParserDockerFormat(testCase.config)

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

func TestParserDockerUnpackingControl(t *testing.T) {
	for _, testCase := range testCasesParserDockerUnpackingControl {
		t.Run(testCase.name, func(t *testing.T) {
			parser := CreateParserDockerFormat(testCase.config)

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

func configFlattenTopLevel() configuration.ParserConfig {
	return configuration.ParserConfig{
		UserLogFieldsKey: "",
		CRIFieldsKey:     "",
		FlattenUserLog:   true,
	}
}

func configFlattenSubDict() configuration.ParserConfig {
	return configuration.ParserConfig{
		UserLogFieldsKey: LogKeyLog,
		CRIFieldsKey:     "cri",
		FlattenUserLog:   true,
	}
}
