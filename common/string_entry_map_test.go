package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntryMapString_Filter(t *testing.T) {
	entryMap := EntryMapString{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	entryMap = entryMap.Filter("key2", "key3", "key_absent")
	assert.Equal(t, EntryMapString{"key2": "value2", "key3": "value3"}, entryMap)

	entryMap = entryMap.Filter("key2")
	assert.Equal(t, EntryMapString{"key2": "value2"}, entryMap)
}

func TestEntryMapString_Extend(t *testing.T) {
	entryMap := EntryMapString{
		"key1": "value1",
		"key2": "value2",
	}

	extend := EntryMapString{
		"key3": "value3",
	}

	entryMap.Extend(extend)
	assert.Equal(t, EntryMapString{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}, entryMap)
}
