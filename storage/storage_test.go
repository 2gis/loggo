package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	TestPath  = "/tmp/test.db"
	TestKey   = "my-key.log"
	TestValue = "value_0"
)

func TestFileRegistry(t *testing.T) {
	storage, err := NewStorage(TestPath, 1)
	assert.NoError(t, err)
	assert.Equal(t, TestPath, storage.path)

	storage.Set(TestKey, TestValue)
	output, err := storage.Get(TestKey)
	assert.NoError(t, err)
	assert.Equal(t, TestValue, output)

	keys, err := storage.Keys()
	assert.NoError(t, err)
	assert.Equal(t, TestKey, keys[0])
	assert.Equal(t, 1, len(keys))

	err = storage.Delete(TestKey)
	assert.NoError(t, err)
	output, err = storage.Get(TestKey)
	assert.NoError(t, err)
	assert.Equal(t, "", output)

	defer os.Remove(TestPath)
}
