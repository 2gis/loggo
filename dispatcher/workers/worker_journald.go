package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/configuration"
	"github.com/2gis/loggo/logging"
)

// workerFollower is an entity which collects log entries obtained by reader and sends them to Out channel
type workerJournald struct {
	worker

	config  configuration.ParserConfig
	extends common.EntryMap

	reader        JournaldReader
	cursorStorage Storage

	sleepNoRecords     time.Duration
	cursorCommitTicker *time.Ticker

	output chan<- string
}

// newFollowerJournald constructor
func newFollowerJournald(output chan<- string, reader JournaldReader, config configuration.ParserConfig,
	extends common.EntryMap, cursorStorage Storage,
	commitIntervalSec, readTimeout int, logger logging.Logger) *workerJournald {
	return &workerJournald{
		worker: worker{
			wg:     &sync.WaitGroup{},
			logger: logger,
		},

		config:  config,
		extends: extends,

		reader:        reader,
		cursorStorage: cursorStorage,

		sleepNoRecords:     time.Duration(readTimeout) * time.Second,
		cursorCommitTicker: time.NewTicker(time.Duration(commitIntervalSec) * time.Second),
		output:             output,
	}
}

// Start starts workerJournald; worker quits when context is done
func (worker *workerJournald) Start(ctx context.Context) {
	worker.wg.Add(2)

	go func() {
		defer worker.wg.Done()
		worker.startCursorCommitter(ctx)
	}()

	go func() {
		defer worker.wg.Done()
		worker.startReader(ctx)
	}()

	worker.wg.Wait()
	worker.finalize()
}

func (worker *workerJournald) startReader(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		worker.Lock()
		err := worker.entryProceed()
		worker.Unlock()

		if err != nil {
			switch err {
			case ErrNoRecords:
				time.Sleep(worker.sleepNoRecords)
			default:
				worker.logger.WithError(err).Warnf("journald encountered error while entry proceed")
			}
		}
	}
}

func (worker *workerJournald) entryProceed() error {
	result := common.EntryMap{}

	entryMap, err := worker.reader.EntryRead()
	if err != nil {
		return err
	}

	if entryMap == nil {
		return ErrNoRecords
	}

	sourceTimestamp, ok := entryMap[common.LabelTime].(string)
	if !ok {
		return fmt.Errorf("missing entry timestamp")
	}

	usec, err := strconv.ParseInt(sourceTimestamp, 10, 64)

	if err != nil {
		return fmt.Errorf("incorrect entry timestamp")
	}

	entryMap["SYSTEMD_UNIT"] = entryMap["_SYSTEMD_UNIT"]
	entryMap = entryMap.Filter(
		"SYSLOG_IDENTIFIER",
		"PRIORITY",
		"SYSLOG_PID",
		"SYSLOG_FACILITY",
		"SYSTEMD_UNIT",
		"MESSAGE",
		common.LabelTime,
	)

	if worker.config.UserLogFieldsKey != "" {
		result[worker.config.UserLogFieldsKey] = entryMap
	} else {
		result = entryMap
	}

	if worker.config.ExtendsFieldsKey != "" {
		result[worker.config.ExtendsFieldsKey] = worker.extends
	} else {
		result.Extend(worker.extends)
	}

	result[common.LabelTime] = time.Unix(0, usec*int64(time.Microsecond)).Format(time.RFC3339)
	entryByteString, err := json.Marshal(result)

	if err != nil {
		return fmt.Errorf("journald: error marshalling entry %v", result)
	}

	worker.output <- string(entryByteString)
	return nil
}

func (worker *workerJournald) startCursorCommitter(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-worker.cursorCommitTicker.C:
			worker.commitCursor()
		}
	}
}

func (worker *workerJournald) finalize() {
	if worker.reader.GetAcquireFlag() {
		worker.commitCursor()
	}

	if err := worker.reader.Close(); err != nil {
		worker.logger.Infof("failed to finalize journal handler")
	}

	worker.cursorCommitTicker.Stop()
}

func (worker *workerJournald) commitCursor() {
	worker.Lock()
	if err := worker.cursorStorage.Set(common.FilenameJournald, worker.reader.GetCursor()); err != nil {
		worker.logger.Info(err)
	}
	worker.Unlock()
}
