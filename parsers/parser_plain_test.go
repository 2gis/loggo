package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/configuration"
)

func TestParseLinePlain(t *testing.T) {
	assert.Equal(t,
		common.EntryMap{"log": common.EntryMap{"msg": "test"}},
		CreateParserPlain(configuration.ParserConfig{
			UserLogFieldsKey: "log",
			RawLogFieldKey:   "msg",
		},
		)([]byte("test")),
	)

	assert.Equal(t,
		common.EntryMap{"log": "test"},
		CreateParserPlain(configuration.ParserConfig{
			UserLogFieldsKey: "",
			RawLogFieldKey:   "log",
		},
		)([]byte("test")),
	)
}
