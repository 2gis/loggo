package parsers

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"

	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/components/k8s"
)

// ErrConstructSLI signals about failed SLA Message construction
var ErrConstructSLI = errors.New(
	"Empty or incorrect field 'request' in message, " +
		"field 'request' should contain method, path and protocol version",
)

type ServiceProvider interface {
	GetServiceByHost(string) *k8s.Service
}

type MetricsCollector interface {
	IncrementHTTPRequestCount(podName, method, service, path, protocol string, status int)
	IncrementHTTPRequestsTotalCount(service string)
	ObserveHTTPRequestTime(podName, method, service, path, protocol string, value float64)
	ObserveHTTPUpstreamResponseTimeTotal(podName, method, service, path, protocol string, value float64)
}

// SLIMessage is a structure for storage the parsed message from MQ
type SLIMessage struct {
	Host    string
	PodName string

	// HTTP request related fields
	Method                    string
	URI                       string
	Protocol                  string
	Status                    int
	RequestTime               float64
	UpstreamResponseTimeTotal *float64
}

// SLI parser modifies metrics according to data received from serviceProvider
type SLI struct {
	serviceProvider  ServiceProvider
	metricsCollector MetricsCollector
}

// NewParserSLI is a constructor for ParserSLI
func NewParserSLI(provider ServiceProvider, collector MetricsCollector) *SLI {
	return &SLI{
		serviceProvider:  provider,
		metricsCollector: collector,
	}
}

// Parse is an interface function to process EntryMap by convention
func (parser *SLI) Parse(entryMap common.EntryMap) {
	slaMessage, err := newSLIMessage(entryMap)

	if err != nil {
		return
	}

	// service check is not called before newSLIMessage checks because there's a lock inside
	service := parser.serviceProvider.GetServiceByHost(slaMessage.Host)

	if service == nil {
		return
	}

	parser.metricsCollector.IncrementHTTPRequestsTotalCount(service.Name)
	pathLabel := service.GetLabelByPath(slaMessage.URI)

	if pathLabel == "" {
		return
	}

	parser.metricsCollector.IncrementHTTPRequestCount(
		slaMessage.PodName,
		slaMessage.Method,
		service.Name,
		pathLabel,
		slaMessage.Protocol,
		slaMessage.Status,
	)
	parser.metricsCollector.ObserveHTTPRequestTime(
		slaMessage.PodName,
		slaMessage.Method,
		service.Name,
		pathLabel,
		slaMessage.Protocol,
		slaMessage.RequestTime,
	)

	if slaMessage.UpstreamResponseTimeTotal == nil {
		return
	}

	parser.metricsCollector.ObserveHTTPUpstreamResponseTimeTotal(
		slaMessage.PodName,
		slaMessage.Method,
		service.Name,
		pathLabel,
		slaMessage.Protocol,
		*slaMessage.UpstreamResponseTimeTotal,
	)
}

// newSLIMessage tries to check if entry map contains service level indicator fields
// type assertions are not so clean and short as previous version's full conversions, but thrice as faster
func newSLIMessage(entryMap common.EntryMap) (sliMessage SLIMessage, err error) {
	// the SLA flag is used to separate the sla messages from regular logs by convention
	slaFlag, _ := entryMap[LogKeySLA].(bool)

	if !slaFlag {
		return SLIMessage{}, ErrConstructSLI
	}

	host, _ := entryMap[LogKeyHost].(string)

	if host == "" {
		return SLIMessage{}, ErrConstructSLI
	}

	method, _ := entryMap[LogKeyRequestMethod].(string)

	if method == "" {
		return SLIMessage{}, ErrConstructSLI
	}

	uri, _ := entryMap[LogKeyRequestURI].(string)

	if uri == "" {
		return SLIMessage{}, ErrConstructSLI
	}

	protocol, _ := entryMap[LogKeyServerProtocol].(string)

	if protocol == "" {
		return SLIMessage{}, ErrConstructSLI
	}

	requestTime, err := strconv.ParseFloat(
		fmt.Sprintf("%v", entryMap[LogKeyRequestTime]), 64,
	)

	if err != nil {
		return SLIMessage{}, ErrConstructSLI
	}

	status, err := strconv.ParseInt(
		fmt.Sprintf("%v", entryMap[LogKeyStatus]), 10, 64,
	)

	if err != nil {
		return SLIMessage{}, ErrConstructSLI
	}

	podName, _ := entryMap[LogKeyUpstreamPodName].(string)
	sliMessage = SLIMessage{
		Host:        host,
		PodName:     podName,
		Method:      method,
		URI:         uri,
		Protocol:    protocol,
		Status:      int(status),
		RequestTime: requestTime,
	}

	if value, ok := entryMap[LogKeyUpstreamResponseTimeTotal]; ok {
		if upstreamResponseTimeTotal, ok := value.(float64); ok {
			sliMessage.UpstreamResponseTimeTotal = &upstreamResponseTimeTotal
		}
	}

	return sliMessage, nil
}
