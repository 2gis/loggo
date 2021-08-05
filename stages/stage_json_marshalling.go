package stages

import (
	"encoding/json"
	"sync"

	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/logging"
)

// StageJSONMarshalling marshals messages from input to JSON and sends them to output
type StageJSONMarshalling struct {
	stage
	input  <-chan common.EntryMap
	output chan string
}

// Out is stage output accessor
func (s *StageJSONMarshalling) Out() <-chan string {
	return s.output
}

// Close closes the stage output after its workers finish
func (s *StageJSONMarshalling) Close() {
	s.stage.Close()
	close(s.output)
}

// NewStageJSONMarshalling is a constructor for StageJSONMarshalling
func NewStageJSONMarshalling(input <-chan common.EntryMap, logger logging.Logger) *StageJSONMarshalling {
	stage := &StageJSONMarshalling{
		stage:  stage{wg: &sync.WaitGroup{}, logger: logger},
		input:  input,
		output: make(chan string),
	}
	stage.stage.proceed = stage.proceed
	return stage
}

func (s *StageJSONMarshalling) proceed() {
	for message := range s.input {
		entry, err := json.Marshal(message)
		if err != nil {
			s.logger.Warnf("Error marshalling log entry %v", message)
			continue
		}

		s.output <- string(entry)
	}
}
