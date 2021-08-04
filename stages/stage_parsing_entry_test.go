package stages

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/logging"
)

func parserFunctionTest(line []byte) (common.EntryMap, error) {
	if len(line) == 0 {
		return common.EntryMap{}, errors.New("line is empty")
	}

	return common.EntryMap{"line": string(line)}, nil
}

func parserFunctionDefaultTest(_ []byte) common.EntryMap {
	return common.EntryMap{
		"default": true,
	}
}

func TestStageParsingEntryTest(t *testing.T) {
	expectations := []struct {
		entry    common.Entry
		entryMap common.EntryMap
	}{
		{
			entry:    common.Entry{Origin: []byte("value")},
			entryMap: common.EntryMap{"line": "value"},
		},
		{
			entry:    common.Entry{Origin: []byte("")},
			entryMap: common.EntryMap{"default": true},
		},
		{
			entry:    common.Entry{Origin: []byte(nil)},
			entryMap: common.EntryMap{"default": true},
		},
	}

	input := make(chan *common.Entry, len(expectations))
	stage := NewStageParsingEntry(input, []ParserFunction{parserFunctionTest}, parserFunctionDefaultTest, logging.NewLoggerDefault())
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		StageInit(stage, 2)
		wg.Done()
	}()

	for _, message := range expectations {
		input <- &message.entry
		assert.Equal(t, <-stage.Out(), message.entryMap)
	}

	close(input)
	wg.Wait()
}
