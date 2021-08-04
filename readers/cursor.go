package readers

import (
	"fmt"
	"strconv"
	"strings"
)

// Cursor position in file
type Cursor struct {
	Inode  uint64
	Device uint64
	Value  int64
}

// String Cursor representation
func (cursor *Cursor) String() string {
	return fmt.Sprintf("%d;%d;%d", cursor.Inode, cursor.Device, cursor.Value)
}

// NewCursorFromString is a constructor
func NewCursorFromString(cursorString string) (*Cursor, error) {
	statInfo := strings.Split(cursorString, ";")

	if len(statInfo) != 3 {
		return &Cursor{}, fmt.Errorf("cannot construct cursor from string %s", cursorString)
	}

	inode, err := strconv.ParseUint(statInfo[0], 10, 64)

	if err != nil {
		return &Cursor{}, fmt.Errorf("cannot construct cursor from string %s", cursorString)
	}

	dev, err := strconv.ParseUint(statInfo[1], 10, 64)

	if err != nil {
		return &Cursor{}, fmt.Errorf("cannot construct cursor from string %s", cursorString)
	}

	value, err := strconv.ParseUint(statInfo[2], 10, 64)

	if err != nil {
		return &Cursor{}, fmt.Errorf("cannot construct cursor from string %s", cursorString)
	}

	return &Cursor{
		Inode:  inode,
		Device: dev,
		Value:  int64(value),
	}, nil
}
