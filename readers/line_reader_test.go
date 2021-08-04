package readers

import (
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	FilePathTemp           = "/tmp/test_line_reader"
	FilePathTempRotated    = "/tmp/test_line_reader_0"
	FilePathTempRegistry   = "/tmp/test_storage"
	ReaderBufferSizeSmall  = 16
	ReaderBufferSizeNormal = 32000
)

func TestLineReader_EntryRead_Positive(t *testing.T) {
	createTestFile([]byte("value1\nvalue2"))

	reader, err := NewLineReader(FilePathTemp, ReaderBufferSizeNormal, &Cursor{}, false)
	assert.NoError(t, err)

	// read entry; second entry is not ended with \n
	byteString, _, err := reader.EntryRead()
	assert.NoError(t, err)

	// check value and cursor
	assert.Equal(t, []byte("value1"), byteString)
	assert.NoError(t, err)
	assert.Equal(t, len("value1\n"), int(reader.GetCursor().Value))

	byteString, _, err = reader.EntryRead()
	assert.NoError(t, err)
	assert.Equal(t, []byte("value2"), byteString)
	assert.Equal(t, len("value1\nvalue2"), int(reader.GetCursor().Value))

	byteString, _, err = reader.EntryRead()
	assert.NoError(t, err)
	assert.Nil(t, byteString)
	assert.Equal(t, len([]byte("value1\nvalue2")), int(reader.GetCursor().Value))

	// append data to file, then check again
	extendTestFile([]byte("value3"))
	byteString, _, err = reader.EntryRead()
	assert.NoError(t, err)
	assert.Equal(t, []byte("value3"), byteString)
	byteString, _, err = reader.EntryRead()
	assert.NoError(t, err)
	assert.Nil(t, byteString)

	assert.Equal(t, len([]byte("value1\nvalue2value3")), int(reader.GetCursor().Value))
	clean()
}

func TestSmallBufferBehavior(t *testing.T) {
	_ = createTestFile([]byte("value1value2value3value4\nvalue5"))

	reader, err := NewLineReader(FilePathTemp, ReaderBufferSizeSmall, &Cursor{}, false)
	assert.NoError(t, err)

	byteString, prefixFlag, err := reader.EntryRead()
	assert.NoError(t, err)

	// check value and cursor
	assert.Equal(t, []byte("value1value2valu"), byteString)
	assert.Equal(t, 16, int(reader.GetCursor().Value))
	assert.True(t, prefixFlag)

	byteString, prefixFlag, err = reader.EntryRead()

	assert.NoError(t, err)
	assert.Equal(t, []byte("e3value4"), byteString)
	assert.Equal(t, int64(25), reader.GetCursor().Value)
	assert.False(t, prefixFlag)

	byteString, prefixFlag, err = reader.EntryRead()
	assert.NoError(t, err)
	assert.Equal(t, []byte("value5"), byteString)
	assert.Equal(t, int64(31), reader.GetCursor().Value)
	assert.False(t, prefixFlag)

	byteString, prefixFlag, err = reader.EntryRead()
	assert.NoError(t, err)
	assert.Equal(t, []byte(nil), byteString)
	assert.Equal(t, int64(31), reader.GetCursor().Value)
	assert.False(t, prefixFlag)
}

func TestLineReader_EntryRead_Positive_Rotation(t *testing.T) {
	_ = createTestFile([]byte("value1\nvalue2"))

	reader, err := NewLineReader(FilePathTemp, ReaderBufferSizeNormal, &Cursor{}, false)
	assert.NoError(t, err)
	rotateTestFile([]byte("value3\nvalue4\n"))

	// underlying buffer still points to previous handler, handler is not refreshed at this point
	reader.EntryRead()
	byteString, _, err := reader.EntryRead()
	assert.NoError(t, err)
	assert.Equal(t, []byte("value2"), byteString)
	assert.Equal(t, len("value1\nvalue2"), int(reader.GetCursor().Value))

	// move counter is 0 here, so it invalidates but returns nil on this iteration
	byteString, _, err = reader.EntryRead()
	assert.NoError(t, err)
	assert.Nil(t, byteString)
	assert.Equal(t, 0, int(reader.GetCursor().Value))

	byteString, _, err = reader.EntryRead()
	assert.NoError(t, err)
	assert.Equal(t, []byte("value3"), byteString)
	assert.Equal(t, len("value3\n"), int(reader.GetCursor().Value))

	byteString, _, err = reader.EntryRead()
	assert.NoError(t, err)
	assert.Equal(t, []byte("value4"), byteString)
	assert.Equal(t, len("value3\nvalue4\n"), int(reader.GetCursor().Value))

	byteString, _, err = reader.EntryRead()
	assert.NoError(t, err)
	assert.Nil(t, byteString)
	clean()
}

func TestLineReader_EntryRead_Negative_Removal(t *testing.T) {
	createTestFile([]byte("value1\nvalue2\n"))

	reader, err := NewLineReader(FilePathTemp, ReaderBufferSizeNormal, &Cursor{}, false)
	assert.NoError(t, err)

	os.Remove(FilePathTemp)

	reader.EntryRead()
	byteString, _, err := reader.EntryRead()
	assert.NoError(t, err)
	assert.Equal(t, []byte("value2"), byteString)

	_, _, err = reader.EntryRead()
	assert.Error(t, err)
	assert.Equal(t, 0, int(reader.GetCursor().Value))

	clean()
}

func TestNewLineReader_EntryRead_Negative_NotAcquired(t *testing.T) {
	clean()
	_, err := NewLineReader(FilePathTemp, ReaderBufferSizeNormal, &Cursor{}, false)
	assert.Error(t, err)
}

func TestLineReader_TailFlag(t *testing.T) {
	createTestFile([]byte("value1\nvalue2"))
	reader, err := NewLineReader(FilePathTemp, ReaderBufferSizeNormal, &Cursor{}, true)
	assert.NoError(t, err)
	extendTestFile([]byte("value3\nvalue4"))

	byteString, _, err := reader.EntryRead()
	assert.NoError(t, err)
	assert.Equal(t, []byte("value2value3"), byteString)

	byteString, _, err = reader.EntryRead()
	assert.NoError(t, err)
	assert.Equal(t, []byte("value4"), byteString)
	clean()
}

func TestLineReader_TailFlag_CursorValid(t *testing.T) {
	createTestFile([]byte("value1\nvalue2"))
	inode, device, _ := getStatInfo()

	// must not consider tail flag if the cursor is valid -- i.e. stat info is the same as in storage record
	reader, err := NewLineReader(FilePathTemp, ReaderBufferSizeNormal, &Cursor{Inode: inode, Device: device, Value: 2}, true)
	assert.NoError(t, err)
	assert.Equal(t, 2, int(reader.GetCursor().Value))
}

func TestLineReader_Close(t *testing.T) {
	createTestFile([]byte("value1\nvalue2"))
	reader, err := NewLineReader(FilePathTemp, ReaderBufferSizeNormal, &Cursor{}, true)
	assert.NoError(t, err)
	reader.Close()

	assert.False(t, reader.GetAcquireFlag())
	_, _, err = reader.EntryRead()
	assert.Error(t, err)
	clean()
}

func TestGetLastSeparatorPosition(t *testing.T) {
	createTestFile([]byte("value1\nvalue2"))
	fileHandler, _ := os.Open(FilePathTemp)
	assert.Equal(t, int64(7), GetLastSeparatorPosition(fileHandler))
	fileHandler.Close()

	createTestFile([]byte("value1value2"))
	fileHandler, _ = os.Open(FilePathTemp)
	assert.Equal(t, int64(0), GetLastSeparatorPosition(fileHandler))
	fileHandler.Close()

	createTestFile([]byte("value1v\n\na\nlue2"))
	fileHandler, _ = os.Open(FilePathTemp)
	assert.Equal(t, int64(11), GetLastSeparatorPosition(fileHandler))
	fileHandler.Close()

	createTestFile([]byte("\n"))
	fileHandler, _ = os.Open(FilePathTemp)
	assert.Equal(t, int64(1), GetLastSeparatorPosition(fileHandler))
	fileHandler.Close()

	createTestFile([]byte(""))
	fileHandler, _ = os.Open(FilePathTemp)
	assert.Equal(t, int64(0), GetLastSeparatorPosition(fileHandler))
	fileHandler.Close()

	createTestFile([]byte("value1\nvalue2"))
	fileHandler, _ = os.Open(FilePathTemp)
	fileHandler.Close()
	assert.Equal(t, int64(0), GetLastSeparatorPosition(fileHandler))
	clean()
}

func getStatInfo() (uint64, uint64, error) {
	stat, err := os.Stat(FilePathTemp)

	if err != nil {
		return 0, 0, err
	}

	statInfo, ok := stat.Sys().(*syscall.Stat_t)

	if !ok {
		return 0, 0, err
	}

	return statInfo.Ino, statInfo.Dev, nil
}

func createTestFile(payload []byte) error {
	file, err := os.Create(FilePathTemp)

	if err != nil {
		return err
	}

	file.Write(payload)
	file.Close()
	return nil
}

func extendTestFile(payload []byte) error {
	file, err := os.OpenFile(FilePathTemp, os.O_APPEND|os.O_RDWR, 0666)

	if err != nil {
		return err
	}

	file.Write(payload)
	file.Close()
	return nil
}

func rotateTestFile(payload []byte) error {
	os.Rename(FilePathTemp, FilePathTempRotated)
	return createTestFile(payload)
}

func clean() {
	os.Remove(FilePathTemp)
	os.Remove(FilePathTempRotated)
	os.Remove(FilePathTempRegistry)
}
