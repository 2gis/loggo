package stages

import (
	"sync"

	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/logging"
)

type ParserFunction func([]byte) (common.EntryMap, error)
type ParserFunctionDefault func([]byte) common.EntryMap

// StageParsingEntry parses lines using given parser, extends resulting map with metadata and sends it downstream
type StageParsingEntry struct {
	stage

	parsers      []ParserFunction
	parseDefault ParserFunctionDefault

	input  <-chan *common.Entry
	output chan common.EntryMap
}

// Out is stage output accessor
func (s *StageParsingEntry) Out() <-chan common.EntryMap {
	return s.output
}

// Close closes the stage output after its workers finish
func (s *StageParsingEntry) Close() {
	s.stage.Close()
	close(s.output)
}

// NewStageParsingEntry is a StageParsingEntry constructor
func NewStageParsingEntry(input <-chan *common.Entry, parsers []ParserFunction,
	parserDefault ParserFunctionDefault, logger logging.Logger) *StageParsingEntry {
	stage := &StageParsingEntry{
		stage: stage{wg: &sync.WaitGroup{}, logger: logger},

		parsers:      parsers,
		parseDefault: parserDefault,

		input:  input,
		output: make(chan common.EntryMap),
	}
	stage.stage.proceed = stage.proceed
	return stage
}

func (s *StageParsingEntry) proceed() {
	for message := range s.input {
		var entryMap common.EntryMap
		var err error

		for _, parse := range s.parsers {
			entryMap, err = parse(message.Origin)

			if err != nil {
				s.logger.Warnf(
					"Error parsing log entry, entry '%s', error '%s'",
					message,
					err,
				)
				continue
			}

			break
		}

		if err != nil {
			entryMap = s.parseDefault(message.Origin)
		}

		entryMap.Extend(message.Extends)
		s.output <- entryMap
	}
}
