package stages

import (
	"errors"
	"sync"

	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/logging"
)

type ParserFunction func([]byte) (common.EntryMap, error)
type ParserFunctionDefault func([]byte) common.EntryMap

var ErrUnknownMessageFormat = errors.New("unknown message log format")

// StageParsingEntry parses lines using given parser, extends resulting map with metadata and sends it downstream
type StageParsingEntry struct {
	stage

	parseDockerFormat     ParserFunction
	parseContainerDFormat ParserFunction
	parseDefault          ParserFunctionDefault

	extendsField string

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
func NewStageParsingEntry(input <-chan *common.Entry, parseDocker, parseContainerD ParserFunction,
	parserDefault ParserFunctionDefault, extendsField string, logger logging.Logger) *StageParsingEntry {
	stage := &StageParsingEntry{
		stage: stage{wg: &sync.WaitGroup{}, logger: logger},

		parseDockerFormat:     parseDocker,
		parseContainerDFormat: parseContainerD,
		parseDefault:          parserDefault,

		extendsField: extendsField,

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

		switch message.Format {
		case common.CRITypeDocker:
			entryMap, err = s.parseDockerFormat(message.Origin)
		case common.CRITypeContainerD:
			entryMap, err = s.parseContainerDFormat(message.Origin)
		default:
			err = ErrUnknownMessageFormat
		}

		if err != nil {
			s.logger.WithField("entry", message).WithError(err).
				Warnf("Error parsing log entry, passing as raw string")
			entryMap = s.parseDefault(message.Origin)
		}

		setExtends(entryMap, message.Extends, s.extendsField)
		s.output <- entryMap
	}
}

func setExtends(entryMap, extends common.EntryMap, extendsField string) {
	if extendsField == "" {
		entryMap.Extend(extends)
		return
	}

	v, ok := entryMap[extendsField].(common.EntryMap)
	if !ok {
		entryMap[extendsField] = extends
		return
	}

	v.Extend(extends)
}
