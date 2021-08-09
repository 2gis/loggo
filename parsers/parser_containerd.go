package parsers

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/2gis/loggo/common"
)

var (
	r = regexp.MustCompile("(?s)^(.+) (stdout|stderr) . (.*)$")
)

// ParseContainerDFormat is a parser for inflated json log record
func ParseContainerDFormat(line []byte) (common.EntryMap, error) {
	lineString := string(line)
	output := r.FindStringSubmatch(lineString)

	if len(output) != containerDLineGroupsCount {
		return nil, fmt.Errorf("unable to parse containerd line '%s'", lineString)
	}

	var outer = make(common.EntryMap)

	outer[LogKeyTime] = output[1]
	outer[LogKeyStream] = output[2]
	outer[LogKeyLog] = output[3]

	var inner interface{}

	err := json.Unmarshal([]byte(output[3]), &inner)
	innerMap, ok := inner.(map[string]interface{})

	if err != nil || !ok {
		return outer, nil
	}

	delete(outer, "log")

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
