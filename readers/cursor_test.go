package readers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCursorFromString(t *testing.T) {
	cursor, err := NewCursorFromString("0;1;2")
	assert.NoError(t, err)
	assert.Equal(t, &Cursor{0, 1, 2}, cursor)
	assert.Equal(t, "0;1;2", cursor.String())
}

func TestNewCursorFromStringNegative(t *testing.T) {
	_, err := NewCursorFromString("-1;1;2")
	assert.Error(t, err)

	_, err = NewCursorFromString("1;-1;2")
	assert.Error(t, err)

	_, err = NewCursorFromString("1;1;-2")
	assert.Error(t, err)

	_, err = NewCursorFromString("1;1;")
	assert.Error(t, err)
	_, err = NewCursorFromString("abc;zxc;1")
	assert.Error(t, err)
}
