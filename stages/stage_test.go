package stages

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

const parallelism = 8

type StageMock struct {
	workersInit int
	wg          *sync.WaitGroup
}

func (s *StageMock) InitWorker() {
	s.wg.Done()
	s.workersInit++
}

func (s *StageMock) Close() {
	s.wg.Wait()
}

func TestStageInit(t *testing.T) {
	stageWg := &sync.WaitGroup{}
	stageWg.Add(parallelism)
	stage := StageMock{wg: stageWg}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		StageInit(&stage, parallelism)
		wg.Done()
	}()

	wg.Wait()
	assert.Equal(t, stage.workersInit, parallelism)
}
