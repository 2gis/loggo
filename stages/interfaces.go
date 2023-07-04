package stages

import "github.com/2gis/loggo/common"

// ParserSLI may inject additional operations on parsed entry
type ParserSLI interface {
	Parse(entryMap common.EntryMap)
}

// MetricsCollector is a metrics counter object interface for parsing sli stage
type MetricsCollector interface {
	IncrementHTTPRequestCount(podName, method, service, path string, status int)
	IncrementHTTPRequestsTotalCount(service string)
	ObserveHTTPRequestTime(podName, method, service, path string, value float64)
	ObserveHTTPUpstreamResponseTimeTotal(podName, method, service, path string, value float64)
}

// TransportClient is a transport interface for transport stage
type TransportClient interface {
	DeliverMessages(messages []string) error
}

// Stage interface
type Stage interface {
	InitWorker()
	Close()
}
