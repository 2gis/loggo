package metrics

import (
	"log"
	"net/http"

	"fmt"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ServeHTTPRequests starts http service for handle metrics
func ServeHTTPRequests(addr string, path string) {
	log.Printf("Start serving metrics on '%s/%s'", addr, path)
	http.Handle(path, promhttp.Handler())
	http.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Ok")
	})
	log.Fatal(http.ListenAndServe(addr, nil))
}

// Collector provides interface for accessing and modification of metrics
type Collector struct {
	httpRequestCount              *prometheus.CounterVec
	httpRequestTotalCount         *prometheus.CounterVec
	httpRequestTime               *prometheus.HistogramVec
	httpUpstreamResponseTimeTotal *prometheus.HistogramVec
	logMessageCount               *prometheus.CounterVec
	throttlingDelay               *prometheus.CounterVec
}

var collector *Collector

// NewCollector is a constructor for Collector singleton
func NewCollector(bucketsString string) (*Collector, error) {
	if collector != nil {
		return collector, nil
	}

	buckets, err := buckets(bucketsString)
	if err != nil {
		return &Collector{}, err
	}

	httpRequestCount := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_request_count",
		Help: "Count requests",
	}, []string{"method", "service", "path", "status", "protocol", "upstream_pod_name"})

	httpRequestTotalCount := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_request_total_count",
		Help: "The total number of requests processed",
	}, []string{"service"})
	httpRequestTime := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_time",
		Help:    "Histogram for HTTP requests time",
		Buckets: buckets,
	}, []string{"method", "service", "path", "protocol", "upstream_pod_name"})
	httpUpstreamResponseTimeTotal := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_upstream_response_time_total",
		Help:    "Histogram for HTTP upstream response time, all the upstreams",
		Buckets: buckets,
	}, []string{"method", "service", "path", "protocol", "upstream_pod_name"})
	logMessageCount := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "log_message_count",
		Help: "Store all processed log messages per one container",
	}, []string{"namespace", "pod", "container"})
	throttlingDelay := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "container_throttling_delay_seconds_total",
		Help: "Indicates particular container's total throttle time",
	}, []string{"namespace", "pod", "container"})

	if err = prometheus.Register(httpRequestCount); err != nil {
		return &Collector{}, err
	}
	if err = prometheus.Register(httpRequestTotalCount); err != nil {
		return &Collector{}, err
	}
	if err = prometheus.Register(httpRequestTime); err != nil {
		return &Collector{}, err
	}
	if err = prometheus.Register(httpUpstreamResponseTimeTotal); err != nil {
		return &Collector{}, err
	}
	if err = prometheus.Register(logMessageCount); err != nil {
		return &Collector{}, err
	}
	if err = prometheus.Register(throttlingDelay); err != nil {
		return &Collector{}, err
	}

	collector = &Collector{
		httpRequestCount:              httpRequestCount,
		httpRequestTotalCount:         httpRequestTotalCount,
		httpRequestTime:               httpRequestTime,
		httpUpstreamResponseTimeTotal: httpUpstreamResponseTimeTotal,
		logMessageCount:               logMessageCount,
		throttlingDelay:               throttlingDelay,
	}
	return collector, nil
}

// Retrieve resets metrics
func (collector *Collector) Retrieve() error {
	collector.httpRequestCount.Reset()
	collector.httpRequestTotalCount.Reset()
	collector.httpRequestTime.Reset()
	collector.httpUpstreamResponseTimeTotal.Reset()
	collector.logMessageCount.Reset()
	collector.throttlingDelay.Reset()
	return nil
}

// IncrementHTTPRequestCount i.golovchenko: don't like this interface actually, probably needs refactoring
func (collector *Collector) IncrementHTTPRequestCount(podName, method, service, path, protocol string, status int) {
	collector.httpRequestCount.With(
		prometheus.Labels{
			"upstream_pod_name": podName,
			"method":            method,
			"service":           service,
			"path":              path,
			"status":            strconv.Itoa(status),
			"protocol":          protocol,
		},
	).Inc()
}

// IncrementHTTPRequestsTotalCount increments corresponding metric
func (collector *Collector) IncrementHTTPRequestsTotalCount(service string) {
	collector.httpRequestTotalCount.With(prometheus.Labels{"service": service}).Inc()
}

// IncrementLogMessageCount increments corresponding metric
func (collector *Collector) IncrementLogMessageCount(
	namespace string, podName string, containerName string) {
	collector.logMessageCount.WithLabelValues(namespace, podName, containerName).Inc()
}

// ObserveHTTPRequestTime should be used to make observations of corresponding metric
func (collector *Collector) ObserveHTTPRequestTime(
	podName, method, service, path, protocol string, value float64) {
	collector.httpRequestTime.With(
		prometheus.Labels{
			"upstream_pod_name": podName,
			"service":           service,
			"method":            method,
			"path":              path,
			"protocol":          protocol,
		},
	).Observe(value)
}

// ObserveHTTPUpstreamResponseTimeTotal should be used to make observations of corresponding metric
func (collector *Collector) ObserveHTTPUpstreamResponseTimeTotal(
	podName, method, service, path, protocol string, value float64) {
	collector.httpUpstreamResponseTimeTotal.With(
		prometheus.Labels{
			"upstream_pod_name": podName,
			"service":           service,
			"method":            method,
			"path":              path,
			"protocol":          protocol,
		},
	).Observe(value)
}

// IncrementThrottlingDelay increments value of corresponding metric
func (collector *Collector) IncrementThrottlingDelay(namespace string, podName string, containerName string, value float64) {
	collector.throttlingDelay.WithLabelValues(namespace, podName, containerName).Add(value)
}

// DeleteThrottlingDelay should be used to delete gauge series with specified labels
func (collector *Collector) DeleteThrottlingDelay(namespace string, podName string, containerName string) bool {
	return collector.throttlingDelay.DeleteLabelValues(namespace, podName, containerName)
}

func buckets(bucketsString string) ([]float64, error) {
	split := strings.Split(bucketsString, " ")
	buckets := make([]float64, 0, len(split))

	for _, bucket := range split {
		parsed, err := strconv.ParseFloat(bucket, 64)

		if err != nil {
			return []float64{}, err
		}

		buckets = append(buckets, parsed)
	}

	return buckets, nil
}
