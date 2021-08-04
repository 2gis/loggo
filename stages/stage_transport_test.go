package stages

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/2gis/loggo/logging"
	"github.com/2gis/loggo/transport/redisclient"
)

func TestStageTransport_FlushByTimeout(t *testing.T) {
	flushInterval := time.Duration(100) * time.Millisecond
	bufferSize := 10

	input := make(chan string, 3)
	redisClient := redisclient.NewClientMock(false)
	stage := NewStageTransport(
		input,
		redisClient,
		bufferSize,
		flushInterval,
		logging.NewLoggerDefault(),
	)
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		StageInit(stage, 2)
		wg.Done()
	}()

	input <- ""
	input <- ""
	input <- ""

	close(input)
	wg.Wait()

	assert.Len(t, redisClient.GetBuffer(), 3)
}

func TestStageTransport_FlushByBufferSize(t *testing.T) {
	flushInterval := time.Duration(1) * time.Hour

	bufferSize := 2
	input := make(chan string)
	redisClient := redisclient.NewClientMock(false)
	stage := NewStageTransport(
		input,
		redisClient,
		bufferSize,
		flushInterval,
		logging.NewLoggerDefault(),
	)
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		StageInit(stage, 1)
		wg.Done()
	}()

	input <- ""
	input <- ""
	input <- ""
	assert.Len(t, redisClient.GetBuffer(), 2)

	input <- ""
	input <- ""
	assert.Len(t, redisClient.GetBuffer(), 4)
	close(input)

	wg.Wait()
	assert.Len(t, redisClient.GetBuffer(), 5)
}
