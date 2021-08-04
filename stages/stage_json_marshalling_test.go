package stages

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/logging"
)

func TestStageJSONMarshalling(t *testing.T) {
	entryMaps := []common.EntryMap{
		{"key_0": "value_0"},
		{"key_1": 1},
		{"key_3": func() {}},
	}

	input := make(chan common.EntryMap, 3)
	stage := NewStageJSONMarshalling(input, logging.NewLoggerDefault())
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		StageInit(stage, 2)
		wg.Done()
	}()

	jsonString, err := json.Marshal(entryMaps[0])
	input <- entryMaps[0]
	assert.Equal(t, string(jsonString), <-stage.Out())
	assert.NoError(t, err)

	jsonString, err = json.Marshal(entryMaps[1])
	input <- entryMaps[1]
	assert.Equal(t, string(jsonString), <-stage.Out())
	assert.NoError(t, err)

	input <- entryMaps[2]
	close(input)
	_, ok := <-stage.Out()
	assert.False(t, ok)

	wg.Wait()
}
