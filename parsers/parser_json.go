package parsers

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/2gis/loggo/common"
)

var errInvalidNginxTimestamp = errors.New("upstream response time key doesn't contain proper value")

// ParseInnerJSON is a parser for inflated json log record
func ParseInnerJSON(line []byte) (common.EntryMap, error) {
	var outer common.EntryMap

	if err := json.Unmarshal(line, &outer); err != nil {
		return nil, fmt.Errorf("outer json unmarshalling, %s", err)
	}

	logFieldContent, ok := outer["log"]

	if !ok {
		return nil, fmt.Errorf("parse error, line '%s', doesn't contain 'log' field", line)
	}

	logFieldContentString, ok := logFieldContent.(string)

	if !ok {
		return nil, fmt.Errorf("parse error, line '%s', log field does not contain string", line)
	}

	var inner interface{}

	err := json.Unmarshal([]byte(logFieldContentString), &inner)
	innerMap, ok := inner.(map[string]interface{})

	if err != nil || !ok {
		outer["log"] = logFieldContentString
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

func nginxUpstreamTimeTransform(value interface{}, total bool) (float64, error) {
	if reflect.TypeOf(value).Kind() == reflect.Float64 {
		return value.(float64), nil
	}

	if reflect.TypeOf(value).Kind() == reflect.String {
		value = strings.Replace(value.(string), " ", "", -1)
		values := strings.Split(value.(string), ",")

		if len(values) == 0 {
			return 0, errInvalidNginxTimestamp
		}

		if !total {
			floatValue, err := strconv.ParseFloat(values[len(values)-1], 64)

			if err != nil {
				return 0, errInvalidNginxTimestamp
			}

			return floatValue, nil
		}

		result := float64(0)
		emptyFlag := true

		for _, value := range values {
			floatValue, err := strconv.ParseFloat(value, 64)

			if err != nil {
				continue
			}

			emptyFlag = false
			result += floatValue
		}

		if emptyFlag {
			return 0, errInvalidNginxTimestamp
		}

		// should we round sum to nearest float with 3 decimal places?
		return math.Round(result*1000) / 1000, nil
	}

	return 0, errInvalidNginxTimestamp
}
