package parsers

import (
	"encoding/json"
	"fmt"

	"github.com/2gis/loggo/common"
)

// ParseDockerFormat is a parser for inflated json log record
func ParseDockerFormat(line []byte) (common.EntryMap, error) {
	var outer common.EntryMap

	if err := json.Unmarshal(line, &outer); err != nil {
		return nil, fmt.Errorf("outer json unmarshalling, %s", err)
	}

	logFieldContent, ok := outer[LogKeyLog]

	if !ok {
		return nil, fmt.Errorf("parse error, line '%s', doesn't contain '%s' field", line, LogKeyLog)
	}

	logFieldContentString, ok := logFieldContent.(string)

	if !ok {
		return nil, fmt.Errorf("parse error, line '%s', '%s' field does not contain string", line, LogKeyLog)
	}

	var inner interface{}

	err := json.Unmarshal([]byte(logFieldContentString), &inner)
	innerMap, ok := inner.(map[string]interface{})

	if err != nil || !ok {
		outer[LogKeyLog] = logFieldContentString
		return outer, nil
	}

	delete(outer, LogKeyLog)

	if err := common.Flatten(outer, innerMap); err != nil {
		return nil, err
	}

	// nginx specific transforms, by convention
	if value, ok := outer[LogKeyUpstreamResponseTime]; ok {
		if transformed, err := nginxUpstreamTimeTransform(value, false); err == nil {
			outer[LogKeyUpstreamResponseTimeReplacement] = transformed
		}

		if transformed, err := nginxUpstreamTimeTransform(value, true); err == nil {
			outer[LogKeyUpstreamResponseTimeTotal] = transformed
		}
	}

	return outer, nil
}
