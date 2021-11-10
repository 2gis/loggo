package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntryMapPlain_Extend(t *testing.T) {
	entryMap := EntryMap{
		"key1": "value1",
		"key2": 0,
	}

	extend := EntryMap{
		"key3": "value3",
	}
	entryMap.Extend(extend)
	assert.Equal(t, EntryMap{
		"key1": "value1",
		"key2": 0,
		"key3": "value3",
	}, entryMap)
}

func TestEntryMap_NamespaceName(t *testing.T) {
	assert.Equal(t, "namespace", EntryMap{KubernetesNamespaceName: "namespace"}.NamespaceName())
	assert.Equal(t, "", EntryMap{}.NamespaceName())
}

func TestEntryMap_PodName(t *testing.T) {
	assert.Equal(t, "pod", EntryMap{KubernetesPodName: "pod"}.PodName())
	assert.Equal(t, "", EntryMap{}.PodName())
}

func TestEntryMap_ContainerName(t *testing.T) {
	assert.Equal(t, "container", EntryMap{KubernetesContainerName: "container"}.ContainerName())
	assert.Equal(t, "", EntryMap{}.ContainerName())
}

func TestEntryMap_Filter(t *testing.T) {
	entryMap := EntryMap{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	entryMap = entryMap.Filter("key2", "key3", "key_absent")
	assert.Equal(t, EntryMap{"key2": "value2", "key3": "value3"}, entryMap)

	entryMap = entryMap.Filter("key2")
	assert.Equal(t, EntryMap{"key2": "value2"}, entryMap)
}
