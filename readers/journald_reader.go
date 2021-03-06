package readers

import (
	"fmt"

	"github.com/coreos/go-systemd/sdjournal"
	"github.com/fsnotify/fsnotify"

	"github.com/2gis/loggo/common"
)

// ReaderJournald is not ReaderJournald in Go stdlib sense. Postponed naming issue.
type ReaderJournald struct {
	journalPath   string
	journal       *sdjournal.Journal
	watcher       *fsnotify.Watcher
	cursor        string
	finalizedFlag bool
}

// NewReaderJournald is a constructor for ReaderJournald instance
func NewReaderJournald(journalPath string, initialCursor string) (*ReaderJournald, error) {
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		return nil, fmt.Errorf("unable to init watching for file '%s', %s", journalPath, err)
	}

	reader := &ReaderJournald{
		journalPath: journalPath,
		journal:     nil,
		cursor:      initialCursor,
		watcher:     watcher,
	}

	err = reader.acquireJournal()

	if err != nil {
		return nil, err
	}

	return reader, nil
}

// EntryRead gets new entry
func (reader *ReaderJournald) EntryRead() (common.EntryMap, error) {
	// journal is absent due to unknown reason
	if !reader.GetAcquireFlag() {
		err := reader.acquireJournal()

		if err != nil {
			return nil, fmt.Errorf("journal from specified path '%s' isn't acquired: %s ", reader.journalPath, err)
		}
	}

	moveCounter, err := reader.journal.Next()

	if err != nil {
		return nil, err
	}

	// check if there any new record
	if moveCounter != 0 {
		cursorNew, err := reader.journal.GetCursor()

		if err != nil {
			return nil, err
		}

		reader.cursor = cursorNew
		entry, err := reader.journal.GetEntry()

		if err != nil {
			return nil, err
		}

		entryMap := make(common.EntryMap, len(entry.Fields))

		for k, v := range entry.Fields {
			entryMap[k] = v
		}

		entryMap[common.LabelTime] = fmt.Sprintf("%d", entry.RealtimeTimestamp)
		return entryMap, nil
	}

	// rotation handling
	select {
	case event := <-reader.watcher.Events:
		if event.Op == fsnotify.Rename || event.Op == fsnotify.Remove {
			err := reader.acquireJournal()
			if err != nil {
				return nil, err
			}

			// successfully rotated, but return empty result in that iteration
			return nil, nil
		}
	default:
	}

	return nil, nil
}

func (reader *ReaderJournald) acquireJournal() error {
	// invalidate journal if it's open
	if reader.journal != nil {
		reader.journal.Close()
		reader.journal = nil
		reader.cursor = ""
	}

	journal, err := sdjournal.NewJournalFromFiles(reader.journalPath)

	if err != nil {
		return err
	}

	reader.journal = journal

	// in case of absent, old or invalid cursor invalidate it
	err = reader.journal.SeekCursor(reader.cursor)

	if err != nil {
		err = reader.journal.SeekHead()
		if err != nil {
			return err
		}

		reader.cursor, _ = reader.journal.GetCursor()
	}

	_ = reader.watcher.Remove(reader.journalPath)

	if err := reader.watcher.Add(reader.journalPath); err != nil {
		return fmt.Errorf("unable to init watching for file '%s', %w", reader.journalPath, err)
	}

	return nil
}

// GetAcquireFlag should be used for checking if reader is ready
func (reader *ReaderJournald) GetAcquireFlag() bool {
	return !reader.finalizedFlag && reader.journal != nil
}

// Close should be used as reader finalizer
func (reader *ReaderJournald) Close() error {
	reader.watcher.Close()
	return reader.journal.Close()
}

// GetCursor should be used to get current reader in-memory cursor;
// journald cursor is not so complex as lineReader's one, so it's just a string
func (reader *ReaderJournald) GetCursor() string {
	return reader.cursor
}
