package stages

import (
	"sync"

	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/logging"
)

// StageParsingSLI checks if the message in its input is a SLI message, modifies metrics, then sends message downstream
type StageParsingSLI struct {
	stage

	parser ParserSLI

	input  <-chan common.EntryMap
	output chan common.EntryMap
}

// Out is stage output accessor
func (s *StageParsingSLI) Out() <-chan common.EntryMap {
	return s.output
}

// Close closes the stage output after its workers finish
func (s *StageParsingSLI) Close() {
	s.stage.Close()
	close(s.output)
}

// NewStageParsingSLI is a StageParsingSLI
func NewStageParsingSLI(input <-chan common.EntryMap, parser ParserSLI, logger logging.Logger) *StageParsingSLI {
	stage := &StageParsingSLI{
		stage:  stage{wg: &sync.WaitGroup{}, logger: logger},
		parser: parser,
		input:  input,
		output: make(chan common.EntryMap),
	}
	stage.stage.proceed = stage.proceed
	return stage
}

func (s *StageParsingSLI) proceed() {
	for message := range s.input {
		s.parser.Parse(message)
		s.output <- message
	}
}
