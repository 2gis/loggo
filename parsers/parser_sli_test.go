package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/2gis/loggo/common"
)

func TestParseCorrectSLAMessage(t *testing.T) {
	source := common.EntryMap{
		LogKeySLA:                             true,
		LogKeyRequestMethod:                   "GET",
		LogKeyRequestURI:                      "/hello",
		LogKeyServerProtocol:                  "1",
		LogKeyHost:                            "test.local",
		LogKeyStatus:                          "200",
		LogKeyRequestTime:                     "10.1",
		LogKeyUpstreamResponseTimeReplacement: 10.2,
		LogKeyUpstreamResponseTimeTotal:       10.2,
	}

	message, err := newSLIMessage(source)
	assert.NoError(t, err)
	assert.Equal(t, getStubMessage("/hello", ""), message)
}

func TestParseCorrectSLAMessageWithQueryParams(t *testing.T) {
	source := common.EntryMap{
		LogKeySLA:                             true,
		LogKeyRequestMethod:                   "GET",
		LogKeyRequestURI:                      "/hello?filter=test",
		LogKeyServerProtocol:                  "1",
		LogKeyHost:                            "test.local",
		LogKeyStatus:                          "200",
		LogKeyRequestTime:                     "10.1",
		LogKeyUpstreamResponseTimeReplacement: 10.2,
		LogKeyUpstreamResponseTimeTotal:       10.2,
	}

	message, err := newSLIMessage(source)
	assert.NoError(t, err)
	assert.Equal(t, getStubMessage("/hello?filter=test", ""), message)
}

func TestParseMessageWithUpstreamPodName(t *testing.T) {
	source := common.EntryMap{
		LogKeySLA:                             true,
		LogKeyRequestMethod:                   "GET",
		LogKeyRequestURI:                      "/hello?filter=test",
		LogKeyServerProtocol:                  "1",
		LogKeyHost:                            "test.local",
		LogKeyStatus:                          "200",
		LogKeyRequestTime:                     "10.1",
		"upstream_pod_name":                   "podname_0",
		LogKeyUpstreamResponseTimeReplacement: 10.2,
		LogKeyUpstreamResponseTimeTotal:       10.2,
	}

	message, err := newSLIMessage(source)
	assert.NoError(t, err)
	assert.Equal(t, getStubMessage("/hello?filter=test", "podname_0"), message)
}

func TestParseMessageWithIncorrectRequest(t *testing.T) {
	source := common.EntryMap{
		LogKeySLA:                             true,
		"request":                             "GET",
		LogKeyHost:                            "test.local",
		LogKeyStatus:                          "200",
		LogKeyRequestTime:                     "10.1",
		LogKeyUpstreamResponseTimeReplacement: 10.2,
		LogKeyUpstreamResponseTimeTotal:       10.2,
	}
	_, err := newSLIMessage(source)
	assert.Error(t, err)

	source = common.EntryMap{
		LogKeyRequestMethod:                   "GET",
		LogKeyHost:                            "test.local",
		LogKeyStatus:                          "200",
		LogKeyRequestTime:                     "10.1",
		LogKeyUpstreamResponseTimeReplacement: 10.2,
		LogKeyUpstreamResponseTimeTotal:       10.2,
	}
	_, err = newSLIMessage(source)
	assert.Error(t, err)

	source = common.EntryMap{
		LogKeyRequestURI:                      "/hello?filter=test",
		LogKeyHost:                            "test.local",
		LogKeyStatus:                          "200",
		LogKeyRequestTime:                     "10.1",
		LogKeyUpstreamResponseTimeReplacement: 10.2,
		LogKeyUpstreamResponseTimeTotal:       10.2,
	}
	_, err = newSLIMessage(source)
	assert.Error(t, err)
}

func getStubMessage(path, podName string) SLIMessage {
	upstreamResponseTime := 10.2
	return SLIMessage{
		URI:                       path,
		Protocol:                  "1",
		Method:                    "GET",
		Status:                    200,
		RequestTime:               10.1,
		UpstreamResponseTimeTotal: &upstreamResponseTime,
		Host:                      "test.local",
		PodName:                   podName,
	}
}
