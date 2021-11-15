package stages

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/logging"
	"github.com/2gis/loggo/parsers"
)

func TestStageFiltering(t *testing.T) {
	inputMessages := []common.EntryMap{
		{},
		{parsers.LogKeyLogging: true, parsers.LogKeySLA: false},
		{parsers.LogKeyLogging: false, parsers.LogKeySLA: true},
	}

	input := make(chan common.EntryMap, len(inputMessages))
	stage := NewStageFiltering(input, "", logging.NewLoggerDefault())
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		StageInit(stage, 2)
		wg.Done()
	}()

	for _, message := range inputMessages {
		input <- message
	}
	close(input)

	outputMessages := make([]common.EntryMap, 0, 2)

	for message := range stage.Out() {
		outputMessages = append(outputMessages, message)
	}

	for _, message := range outputMessages {
		_, ok := message[parsers.LogKeyLogging]
		assert.False(t, ok)
		_, ok = message[parsers.LogKeySLA]
		assert.False(t, ok)
	}

	assert.Len(t, outputMessages, 2)
	wg.Wait()
}

func TestStageFilteringNestedField(t *testing.T) {
	inputMessages := []common.EntryMap{
		{"log": "test"},
		{"log": common.EntryMap{
			parsers.LogKeyLogging: true, parsers.LogKeySLA: false,
		}},
		{"log": common.EntryMap{
			parsers.LogKeyLogging: false, parsers.LogKeySLA: true,
		}},
	}

	input := make(chan common.EntryMap, len(inputMessages))
	stage := NewStageFiltering(input, "log", logging.NewLoggerDefault())
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		StageInit(stage, 2)
		wg.Done()
	}()

	for _, message := range inputMessages {
		input <- message
	}
	close(input)

	outputMessages := make([]common.EntryMap, 0, 2)

	for message := range stage.Out() {
		outputMessages = append(outputMessages, message)
	}

	for _, message := range outputMessages {
		_, ok := message[parsers.LogKeyLogging]
		assert.False(t, ok)
		_, ok = message[parsers.LogKeySLA]
		assert.False(t, ok)
	}

	assert.Len(t, outputMessages, 2)
	wg.Wait()
}
