package workers

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/2gis/loggo/components/rates"
	"github.com/2gis/loggo/tests/mocks"

	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/logging"
	"github.com/2gis/loggo/readers"
	"github.com/2gis/loggo/storage"
)

const (
	FilePathTemp           = "/tmp/test_line_reader"
	FilePathTempRegistry   = "/tmp/test_storage"
	readRate               = 1000
	commitInterval         = 1
	limiterUpdateInterval  = 1
	sleepNoRecordsInterval = 1
)

func TestFollower_Positive(t *testing.T) {
	output := make(chan *common.Entry)
	ctx, _ := context.WithCancel(context.Background())
	cursorStorage, err := storage.NewStorage(FilePathTempRegistry, 10)
	assert.NoError(t, err)

	lines := [][]byte{
		[]byte(`value_0`),
		[]byte(`value_1`),
		[]byte(`value_2`),
	}

	reader := mocks.NewLineReaderMock(lines...)
	extends := common.EntryMap{"extend": "extend_value"}
	wg := &sync.WaitGroup{}
	rater, err := rates.NewRater(rates.NewRuleRecordsProviderStub(), readRate)
	assert.NoError(t, err)

	follower := newFollower(
		output,
		FilePathTemp,
		reader,
		mocks.NewCollectorMock(),
		cursorStorage,
		rater,
		extends,
		sleepNoRecordsInterval,
		commitInterval,
		limiterUpdateInterval,
		logging.NewLoggerDefault(),
	)
	wg.Add(1)

	go func() {
		follower.Start(ctx)
		wg.Done()
	}()

	for i := 0; i < len(lines); i++ {
		assert.Equal(t, &common.Entry{Origin: lines[i], Extends: extends}, <-output)
	}

	follower.Stop()
	wg.Wait()
	close(output)
	_ = os.Remove(FilePathTempRegistry)
}

func TestFollowerBase_ShutdownOnFileMissing(t *testing.T) {
	output := make(chan *common.Entry)
	ctx, stop := context.WithCancel(context.Background())
	cursorStorage, err := storage.NewStorage(FilePathTempRegistry, 10)
	assert.NoError(t, err)
	reader := mocks.NewLineReaderMock()
	wg := &sync.WaitGroup{}
	rater, err := rates.NewRater(rates.NewRuleRecordsProviderStub(), readRate)
	assert.NoError(t, err)

	follower := newFollower(
		output,
		FilePathTemp,
		reader,
		mocks.NewCollectorMock(),
		cursorStorage,
		rater,
		common.EntryMap{},
		sleepNoRecordsInterval,
		commitInterval,
		limiterUpdateInterval,
		logging.NewLoggerDefault(),
	)
	wg.Add(1)

	go func() {
		follower.Start(ctx)
		wg.Done()
	}()

	// must shutdown on removals
	reader.SetReturnErrorFlag(true)
	stop()
	wg.Wait()
	assert.False(t, follower.GetActiveFlag())
	_ = os.Remove(FilePathTempRegistry)
}

func TestFollowerBase_CommitCursorOnStop(t *testing.T) {
	output := make(chan *common.Entry)
	ctx, stop := context.WithCancel(context.Background())
	cursorStorage, err := storage.NewStorage(FilePathTempRegistry, 10)
	assert.NoError(t, err)
	reader := mocks.NewLineReaderMock()
	wg := &sync.WaitGroup{}
	rater, err := rates.NewRater(rates.NewRuleRecordsProviderStub(), readRate)
	assert.NoError(t, err)

	follower := newFollower(
		output,
		FilePathTemp,
		reader,
		mocks.NewCollectorMock(),
		cursorStorage,
		rater,
		common.EntryMap{},
		sleepNoRecordsInterval,
		commitInterval,
		limiterUpdateInterval,
		logging.NewLoggerDefault(),
	)
	wg.Add(1)

	go func() {
		follower.Start(ctx)
		wg.Done()
	}()

	// set cursor to an arbitrary value, then check it's committed
	testCursorValue := int64(42)
	reader.SetCursor(&readers.Cursor{Value: testCursorValue})
	stop()
	wg.Wait()

	value, err := cursorStorage.Get(FilePathTemp)
	assert.NoError(t, err)
	cursor, err := readers.NewCursorFromString(value)
	assert.NoError(t, err)
	assert.Equal(t, testCursorValue, cursor.Value)
	_ = os.Remove(FilePathTempRegistry)

	// must not commit from reader that hasn't acquired file
	ctx, stop = context.WithCancel(context.Background())
	cursorStorage, err = storage.NewStorage(FilePathTempRegistry, 10)
	assert.NoError(t, err)
	reader = mocks.NewLineReaderMock()

	follower = newFollower(
		output,
		FilePathTemp,
		reader,
		mocks.NewCollectorMock(),
		cursorStorage,
		rater,
		common.EntryMap{},
		sleepNoRecordsInterval,
		commitInterval,
		limiterUpdateInterval,
		logging.NewLoggerDefault(),
	)
	wg.Add(1)

	go func() {
		follower.Start(ctx)
		wg.Done()
	}()

	reader.SetAcquireFlag(false)
	reader.SetCursor(&readers.Cursor{Value: testCursorValue})
	stop()
	wg.Wait()

	value, err = cursorStorage.Get(FilePathTemp)
	assert.NoError(t, err)
	assert.Equal(t, "", value)
	_ = os.Remove(FilePathTempRegistry)
}
