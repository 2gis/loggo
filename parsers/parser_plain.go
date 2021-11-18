package parsers

import (
	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/configuration"
)

func CreateParserPlain(config configuration.ParserConfig) func(line []byte) common.EntryMap {
	return func(line []byte) common.EntryMap {
		result := common.EntryMap{}
		baseMap := selectBaseMap(result, config.UserLogFieldsKey)
		baseMap[config.RawLogFieldKey] = string(line)

		return result
	}
}
