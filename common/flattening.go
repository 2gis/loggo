package common

import (
	"errors"
	"fmt"
)

// ErrInvalidInput signals about invalid input to Flatten
var ErrInvalidInput = errors.New("not a valid input: map expected")

// Flatten unpacks nested map into top map.
// Keys in the flat map will be a compound of descending map keys and slice iterations
func Flatten(top map[string]interface{}, nested map[string]interface{}) error {
	return flatten(true, top, nested, "")
}

func flatten(top bool, flatMap map[string]interface{}, nested interface{}, prefix string) error {
	switch nested.(type) {
	case map[string]interface{}:
		for key, value := range nested.(map[string]interface{}) {
			newKey := enkey(top, prefix, key)

			switch value.(type) {
			case map[string]interface{}:
				if err := flatten(false, flatMap, value, newKey); err != nil {
					return err
				}
			default:
				flatMap[newKey] = value
			}
		}
	default:
		return ErrInvalidInput
	}

	return nil
}

func enkey(top bool, prefix, subkey string) string {
	if top {
		return fmt.Sprintf("%s%s", prefix, subkey)
	}

	return fmt.Sprintf("%s.%s", prefix, subkey)
}
