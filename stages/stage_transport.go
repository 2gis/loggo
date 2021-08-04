package stages

import (
	"sync"
	"time"

	"github.com/2gis/loggo/logging"
)

// StageTransport stacks messages from input into batches and tries to send them to transport
type StageTransport struct {
	stage

	transportClient TransportClient

	bufferSizeMax int
	flushInterval time.Duration

	input <-chan string
}

// Close waits its workers finish
func (s *StageTransport) Close() {
	s.stage.Close()
}

// NewStageTransport is a StageTransport constructor
func NewStageTransport(input <-chan string, transportClient TransportClient,
	bufferSizeMax int, flushInterval time.Duration, logger logging.Logger) *StageTransport {
	stage := &StageTransport{
		stage: stage{wg: &sync.WaitGroup{}, logger: logger},

		transportClient: transportClient,

		bufferSizeMax: bufferSizeMax,
		flushInterval: flushInterval,

		input: input,
	}
	stage.stage.proceed = stage.proceed
	return stage
}

func (s *StageTransport) proceed() {
	ticker := time.NewTicker(s.flushInterval)
	buffer := make([]string, 0, s.bufferSizeMax)

	defer ticker.Stop()

	for {
		if len(buffer) >= s.bufferSizeMax {
			flushedCount, err := s.flush(&buffer, false)

			if err != nil {
				s.logger.Errorf(
					"failed flushing buffer by buffer size, error '%v', keeping current buffer with %d records",
					err,
					flushedCount,
				)
				time.Sleep(SleepTransportUnavailable)
				continue
			}
		}

		select {
		case <-ticker.C:
			flushedCount, err := s.flush(&buffer, false)

			if err != nil {
				s.logger.Errorf(
					"failed flushing buffer by timeout, error '%v', keeping current buffer with %d records",
					err,
					flushedCount,
				)
				time.Sleep(SleepTransportUnavailable)
				continue
			}

		case data, ok := <-s.input:
			if !ok {
				if flushedCount, err := s.flush(&buffer, true); err != nil {
					s.logger.Errorf("failed flushing buffer while shutting down, "+
						"error '%v', current buffer with size %d records will be lost",
						err,
						flushedCount,
					)
				}
				return
			}

			buffer = append(buffer, data)
		}
	}
}

func (s *StageTransport) flush(buffer *[]string, forceFlag bool) (bufferSizeOld int, err error) {
	bufferSizeOld = len(*buffer)
	if bufferSizeOld == 0 {
		return
	}

	if err = s.transportClient.DeliverMessages(*buffer); err != nil && !forceFlag {
		return
	}

	*buffer = make([]string, 0, s.bufferSizeMax)
	return
}
