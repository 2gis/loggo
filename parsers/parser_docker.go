package parsers

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/configuration"
)

var (
	ErrLogFieldMissing   = errors.New("line does not contain user log field")
	ErrLogFieldNotString = errors.New("user log field does not contain string")
)

func CreateParserDockerFormat(config configuration.ParserConfig) func(line []byte) (common.EntryMap, error) {
	return func(line []byte) (common.EntryMap, error) {
		var outer common.EntryMap

		if err := json.Unmarshal(line, &outer); err != nil {
			return nil, fmt.Errorf("outer json unmarshalling, %s", err)
		}

		logFieldContent, ok := outer[LogKeyLog]

		if !ok {
			return nil, fmt.Errorf("%w: line '%s'", ErrLogFieldMissing, line)
		}

		logFieldContentString, ok := logFieldContent.(string)

		if !ok {
			return nil, fmt.Errorf("%w: line '%s'", ErrLogFieldNotString, line)
		}

		delete(outer, LogKeyLog)
		setDockerFields(outer, config.CRIFieldsKey)

		if err := setLogFieldContent(
			outer, config.UserLogFieldsKey, config.RawLogFieldKey, logFieldContentString, config.FlattenUserLog); err != nil {
			return nil, fmt.Errorf("error setting user log field: %w", err)
		}

		return outer, nil
	}
}

func setDockerFields(entryMap common.EntryMap, targetField string) {
	if targetField == "" {
		return
	}

	if len(entryMap) == 0 {
		return
	}

	dockerFields := make(common.EntryMap)

	for k, v := range entryMap {
		dockerFields[k] = v
		delete(entryMap, k)
	}

	entryMap[targetField] = dockerFields
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
