package stages

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/logging"
	"github.com/2gis/loggo/tests/mocks"
)

func TestStageParsingSLITest(t *testing.T) {
	input := make(chan common.EntryMap)
	parser := mocks.NewSLIMock()
	stage := NewStageParsingSLI(input, parser, logging.NewLoggerDefault())
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		StageInit(stage, 2)
		wg.Done()
	}()

	input <- common.EntryMap{}
	close(input)
	<-stage.Out()

	wg.Wait()
	assert.Equal(t, 1, parser.Parsed())
}
