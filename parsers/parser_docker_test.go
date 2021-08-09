package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/2gis/loggo/common"
)

func TestParseLine(t *testing.T) {
	// plain positive
	out, err := ParseDockerFormat([]byte("{\"log\":\"hello world\", \"key1\":1}"))
	assert.NoError(t, err)
	assert.Equal(t, common.EntryMap{"log": "hello world", "key1": float64(1)}, out)

	// key "log" is absent
	_, err = ParseDockerFormat([]byte("{\"somekey\":\"hello world\"}"))
	assert.Error(t, err)

	// plain negative (bad json syntax)
	_, err = ParseDockerFormat([]byte("{log\":\"hello world\"}"))
	assert.Error(t, err)

	// inner map positive. Can't assert for object directly cause interface{}(nil) is not equal with itself
	out, err = ParseDockerFormat([]byte(`{"log": "{\"hello\":\"world\",\"a\": 1,\"b\": null}"}`))
	assert.NoError(t, err)
	assert.Equal(t, "world", out["hello"])
	assert.Equal(t, float64(1), out["a"])
	assert.Nil(t, out["b"])
}

func TestParseUpstreamResponseTime(t *testing.T) {
	out, err := ParseDockerFormat([]byte(`{
		"log": "{\"upstream_response_time\":\"-\"}"
		}`))
	assert.NoError(t, err)
	assert.Equal(t, "-", out[LogKeyUpstreamResponseTime])
	assert.Equal(t, nil, out[LogKeyUpstreamResponseTimeReplacement])

	out, err = ParseDockerFormat([]byte(`{
		"log": "{\"upstream_response_time\":\"0\"}"
		}`))
	assert.NoError(t, err)
	assert.Equal(t, "0", out[LogKeyUpstreamResponseTime])
	assert.Equal(t, float64(0), out[LogKeyUpstreamResponseTimeReplacement])

	out, err = ParseDockerFormat([]byte(`{
		"log": "{\"upstream_response_time\":\"0.009\"}"
		}`))
	assert.NoError(t, err)
	assert.Equal(t, "0.009", out[LogKeyUpstreamResponseTime])
	assert.Equal(t, float64(0.009), out[LogKeyUpstreamResponseTimeReplacement])

	out, err = ParseDockerFormat([]byte(`{
		"log": "{\"upstream_response_time\":\"0.009,1.142\"}"
		}`))
	assert.NoError(t, err)
	assert.Equal(t, "0.009,1.142", out[LogKeyUpstreamResponseTime])
	assert.Equal(t, float64(1.142), out[LogKeyUpstreamResponseTimeReplacement])
	assert.Equal(t, float64(1.151), out[LogKeyUpstreamResponseTimeTotal])

	out, err = ParseDockerFormat([]byte(`{
		"log": "{\"upstream_response_time\":\"0.009, 1.142, 1.222\"}"
		}`))
	assert.NoError(t, err)
	assert.Equal(t, "0.009, 1.142, 1.222", out[LogKeyUpstreamResponseTime])
	assert.Equal(t, float64(1.222), out[LogKeyUpstreamResponseTimeReplacement])
	assert.Equal(t, float64(2.373), out[LogKeyUpstreamResponseTimeTotal])

	// maybe convert to string anyway?..
	out, err = ParseDockerFormat([]byte(`{
		"log": "{\"upstream_response_time\":1.222}"
		}`))
	assert.NoError(t, err)
	assert.Equal(t, float64(1.222), out[LogKeyUpstreamResponseTime])
	assert.Equal(t, float64(1.222), out[LogKeyUpstreamResponseTimeReplacement])

	out, err = ParseDockerFormat([]byte(`{
		"log": "{\"upstream_response_time\":11}"
		}`))
	assert.NoError(t, err)
	assert.Equal(t, float64(11), out[LogKeyUpstreamResponseTime])
	assert.Equal(t, float64(11), out[LogKeyUpstreamResponseTimeReplacement])

	out, err = ParseDockerFormat([]byte(`{
		"log": "{\"upstream_response_time\":[1.142, 1.222]}"
		}`))
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{1.142, 1.222}, out[LogKeyUpstreamResponseTime])
	assert.Nil(t, out[LogKeyUpstreamResponseTimeReplacement])

	out, err = ParseDockerFormat([]byte(`{
		"log": "{\"upstream_response_time\":{\"value\": 1.142}}"
		}`))
	assert.NoError(t, err)
	assert.Equal(t, float64(1.142), out[LogKeyUpstreamResponseTime+".value"])
}
