package stages

import (
	"sync"

	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/logging"
	"github.com/2gis/loggo/parsers"
)

// StageFiltering filters messages from input
type StageFiltering struct {
	stage
	userLogField string

	input  <-chan common.EntryMap
	output chan common.EntryMap
}

// Out is stage output accessor
func (s *StageFiltering) Out() <-chan common.EntryMap {
	return s.output
}

// Close closes the stage output after its workers finish
func (s *StageFiltering) Close() {
	s.stage.Close()
	close(s.output)
}

// NewStageFiltering is a StageFiltering constructor
func NewStageFiltering(input <-chan common.EntryMap, userLogField string, logger logging.Logger) *StageFiltering {
	stage := &StageFiltering{
		stage:        stage{wg: &sync.WaitGroup{}, logger: logger},
		userLogField: userLogField,
		input:        input,
		output:       make(chan common.EntryMap),
	}
	stage.stage.proceed = stage.proceed
	return stage
}

func (s *StageFiltering) proceed() {
	for message := range s.input {
		if s.userLogField == "" {
			if !filterOutSerivceFields(message) {
				continue
			}
		}

		if v, ok := message[s.userLogField].(map[string]interface{}); ok {
			if !filterOutSerivceFields(v) {
				continue
			}
		}

		s.output <- message
	}
}

func filterOutSerivceFields(message map[string]interface{}) bool {
	loggingFlag := true

	if flag, ok := message[parsers.LogKeyLogging].(bool); ok {
		loggingFlag = flag
	}

	delete(message, parsers.LogKeyLogging)
	delete(message, parsers.LogKeySLA)

	return loggingFlag
}
