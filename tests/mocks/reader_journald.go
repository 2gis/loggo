package mocks

import (
	"errors"
	"fmt"

	"github.com/2gis/loggo/common"
)

var ErrorClosed = errors.New("not acquired")

// ReaderJournaldMock mocks specific sdjournal based reader with EntryMap reader interface
type ReaderJournaldMock struct {
	Entries       []common.EntryMapString
	cursor        int
	finalizedFlag bool
}

func NewReaderJournaldMock(entries ...common.EntryMapString) *ReaderJournaldMock {
	return &ReaderJournaldMock{
		Entries: entries,
	}
}

func (reader *ReaderJournaldMock) GetAcquireFlag() bool {
	return !reader.finalizedFlag
}

func (reader *ReaderJournaldMock) GetCursor() string {
	return fmt.Sprintf("%d", reader.cursor)
}

func (reader *ReaderJournaldMock) Close() error {
	if reader.finalizedFlag {
		return ErrorClosed
	}

	reader.finalizedFlag = true
	return nil
}

func (reader *ReaderJournaldMock) EntryRead() (common.EntryMapString, error) {
	if reader.finalizedFlag {
		return nil, ErrorClosed
	}

	if reader.cursor > len(reader.Entries)-1 {
		return nil, nil
	}

	result := reader.Entries[reader.cursor]
	reader.cursor++
	return result, nil
}
