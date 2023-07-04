package parsers

// LogKeyLog field that contains actual user log in log line.
const LogKeyLog = "log"

/* Nginx related particular log keys */
const (
	LogKeyUpstreamResponseTime            = "upstream_response_time"
	LogKeyUpstreamResponseTimeReplacement = "upstream_response_time_float"
	LogKeyUpstreamResponseTimeTotal       = "upstream_response_time_total"
	LogKeyHost                            = "host"
	LogKeyRequestMethod                   = "request_method"
	LogKeyRequestURI                      = "request_uri"
	LogKeyRequestTime                     = "request_time"
	LogKeyStatus                          = "status"
	LogKeyUpstreamPodName                 = "upstream_pod_name"
)

/* Processing control related fields */
const (
	LogKeySLA     = "sla"
	LogKeyLogging = "logging"
)

/* ContainerD related constants */
const (
	containerDLineGroupsCount = 4
	LogKeyTime                = "time"
	LogKeyStream              = "stream"
)
