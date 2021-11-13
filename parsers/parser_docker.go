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
			outer, config.UserLogFieldsKey, logFieldContentString, config.FlattenUserLog); err != nil {
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

func setLogFieldContent(entryMap common.EntryMap, targetField, JSONContent string, flatten bool) error {
	var inner interface{}

	err := json.Unmarshal([]byte(JSONContent), &inner)
	innerMap, ok := inner.(map[string]interface{})

	if err != nil || !ok {
		if targetField == "" {
			entryMap[LogKeyLog] = JSONContent
			return nil
		}

		entryMap[targetField] = JSONContent
		return nil
	}

	processNginxFields(innerMap)

	if !flatten {
		// flatten flag is not set and the target field is empty, we still need some field to set content to.
		if targetField == "" {
			entryMap[LogKeyLog] = common.EntryMap(innerMap)
			return nil
		}

		entryMap[targetField] = common.EntryMap(innerMap)
		return nil
	}

	// unpack to top level dict, backward compatibility
	if targetField == "" {
		return common.Flatten(entryMap, innerMap)
	}

	innerMapUnpacked := make(common.EntryMap)

	if err := common.Flatten(innerMapUnpacked, innerMap); err != nil {
		return err
	}

	entryMap[targetField] = innerMapUnpacked
	return nil
}
