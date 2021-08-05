package parsers

/* Nginx particular log keys */
const (
	LogKeyUpstreamResponseTime            = "upstream_response_time"
	LogKeyUpstreamResponseTimeReplacement = "upstream_response_time_float"
	LogKeyUpstreamResponseTimeTotal       = "upstream_response_time_total"
	LogKeyHost                            = "host"
	LogKeyRequestMethod                   = "request_method"
	LogKeyRequestURI                      = "request_uri"
	LogKeyServerProtocol                  = "server_protocol"
	LogKeyRequestTime                     = "request_time"
	LogKeyStatus                          = "status"
	LogKeyUpstreamPodName                 = "upstream_pod_name"
	LogKeySLA                             = "sla"
	LogKeyLogging                         = "logging"
)
