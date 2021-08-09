package parsers

import (
	"errors"
	"math"
	"strconv"
	"strings"
)

var errInvalidNginxTimestamp = errors.New("upstream response time key doesn't contain proper value")

func nginxUpstreamTimeTransform(value interface{}, total bool) (float64, error) {
	if v, ok := value.(float64); ok {
		return v, nil
	}

	if v, ok := value.(string); ok {
		return nginxUpstreamTimeFromString(v, total)
	}

	return 0, errInvalidNginxTimestamp
}

func nginxUpstreamTimeFromString(v string, total bool) (float64, error) {
	v = strings.Replace(v, " ", "", -1)
	values := strings.Split(v, ",")

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
