package common

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestFlatten(t *testing.T) {
	cases := []struct {
		test string
		want EntryMap
	}{
		{
			`{
				"foo": {
					"jim": "bean",
					"something": {
						"nested_key": "nested_value"
					}
				},
				"fee": "bar",
				"n1": {
					"alist": [
						"a",
						"b",
						"c",
						{
							"d": "other",
							"e": "another"
						}
					]
				},
				"number": 1.4567,
				"bool": true
			}`,
			EntryMap{
				"foo.jim":                  "bean",
				"fee":                      "bar",
				"foo.something.nested_key": "nested_value",
				"n1.alist":                 []interface{}{"a", "b", "c", map[string]interface{}{"d": "other", "e": "another"}},
				"number":                   1.4567,
				"bool":                     true,
			},
		},
	}

	for i, test := range cases {
		var m interface{}
		err := json.Unmarshal([]byte(test.test), &m)
		if err != nil {
			t.Errorf("%d: failed to unmarshal test: %v", i+1, err)
			continue
		}
		top := make(EntryMap)
		err = Flatten(top, m.(map[string]interface{}))
		if err != nil {
			t.Errorf("%d: failed to flatten: %v", i+1, err)
			continue
		}
		if !reflect.DeepEqual(top, test.want) {
			t.Errorf("%d: mismatch, got: %v wanted: %v", i+1, top, test.want)
		}
	}
}
