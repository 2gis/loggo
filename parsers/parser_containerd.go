package parsers

import (
	"fmt"
	"regexp"

	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/configuration"
)

var (
	r = regexp.MustCompile("(?s)^(.+) (stdout|stderr) . (.*)$")
)

// CreateParserContainerDFormat returns containerd parser
func CreateParserContainerDFormat(config configuration.ParserConfig) func(line []byte) (common.EntryMap, error) {
	return func(line []byte) (common.EntryMap, error) {
		lineString := string(line)
		output := r.FindStringSubmatch(lineString)

		if len(output) != containerDLineGroupsCount {
			return nil, fmt.Errorf("unable to parse containerd line '%s'", lineString)
		}

		var outer = make(common.EntryMap)

		setContainerDFields(outer, config.CRIFieldsKey, output[1], output[2])

		if err := setLogFieldContent(outer, config.UserLogFieldsKey, output[3], config.FlattenUserLog); err != nil {
			return nil, fmt.Errorf("error setting user log field: %w", err)
		}

		return outer, nil
	}
}

func setContainerDFields(entryMap common.EntryMap, targetField, time, stream string) {
	if targetField == "" {
		entryMap[LogKeyTime] = time
		entryMap[LogKeyStream] = stream
		return
	}

	fields := make(map[string]interface{})
	fields[LogKeyTime] = time
	fields[LogKeyStream] = stream
	entryMap[targetField] = fields
}
