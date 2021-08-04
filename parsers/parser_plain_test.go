package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/2gis/loggo/common"
)

func TestParseLinePlain(t *testing.T) {
	assert.Equal(t, common.EntryMap{"log": "test"}, ParseStringPlain([]byte("test")))
}
