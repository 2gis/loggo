package parsers

import (
	"encoding/json"
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

		if err := setLogFieldContent(
			outer, config.UserLogFieldsKey, config.RawLogFieldKey, output[3], config.FlattenUserLog); err != nil {
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

	entryMap[targetField] = common.EntryMap{
		LogKeyTime:   time,
		LogKeyStream: stream,
	}
}

func setLogFieldContent(entryMap common.EntryMap, userLogField, rawField, logFieldContent string, flatten bool) error {
	var inner interface{}

	baseMap := selectBaseMap(entryMap, userLogField)
	err := json.Unmarshal([]byte(logFieldContent), &inner)
	innerMap, ok := inner.(map[string]interface{})
	if err != nil || !ok {
		baseMap[rawField] = logFieldContent
		return nil
	}

	processNginxFields(innerMap)

	if !flatten {
		baseMap.Extend(innerMap)
		return nil
	}

	return common.Flatten(baseMap, innerMap)
}

func selectBaseMap(baseMap common.EntryMap, userLogField string) common.EntryMap {
	if userLogField == "" {
		return baseMap
	}

	subMap := make(common.EntryMap)
	baseMap[userLogField] = subMap
	return subMap
}
