package mocks

import (
	"errors"

	"github.com/2gis/loggo/readers"
)

// LineReaderMock is mock for readers with byte ReaderJournald interface
type LineReaderMock struct {
	Entries         [][]byte
	cursor          *readers.Cursor
	acquireFlag     bool
	prefixFlag      bool
	returnErrorFlag bool
}

func NewLineReaderMock(entries ...[]byte) *LineReaderMock {
	return &LineReaderMock{
		Entries:     entries,
		cursor:      &readers.Cursor{},
		acquireFlag: true,
	}
}

func (reader *LineReaderMock) SetPrefixFlag(flag bool) {
	reader.prefixFlag = flag
}

func (reader *LineReaderMock) GetPrefixFlag() bool {
	return reader.prefixFlag
}

func (reader *LineReaderMock) SetAcquireFlag(flag bool) {
	reader.acquireFlag = flag
}

func (reader *LineReaderMock) GetAcquireFlag() bool {
	return reader.acquireFlag
}

func (reader *LineReaderMock) Close() error {
	return nil
}

func (reader *LineReaderMock) EntryRead() ([]byte, bool, error) {

	if !reader.acquireFlag {
		return nil, reader.prefixFlag, errors.New("any error")
	}

	if reader.returnErrorFlag {
		return nil, reader.prefixFlag, errors.New("any error")
	}

	if int(reader.cursor.Value) > len(reader.Entries)-1 {
		return nil, reader.prefixFlag, nil
	}

	result := reader.Entries[reader.cursor.Value]
	reader.cursor.Value++
	return result, reader.prefixFlag, nil
}

func (reader *LineReaderMock) SetCursor(cursor *readers.Cursor) {
	reader.cursor = cursor
}

// GetCursor should be used to get current reader in-memory cursor
func (reader *LineReaderMock) GetCursor() *readers.Cursor {
	return reader.cursor
}

func (reader *LineReaderMock) SetReturnErrorFlag(flag bool) {
	reader.returnErrorFlag = flag
}
