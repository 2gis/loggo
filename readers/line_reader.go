package readers

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"syscall"
)

// LineReader is used for reading lines from file. Signals about rotations (fs notify rename)
type LineReader struct {
	filePath     string
	fileHandler  *os.File
	buffer       *bufio.Reader
	bufferSize   int
	cursor       *Cursor
	fromTailFlag bool
}

// NewLineReader is a constructor for LineReader
func NewLineReader(filePath string, bufferSize int, initialCursor *Cursor, fromTailFlag bool) (*LineReader, error) {
	reader := LineReader{
		filePath:     filePath,
		bufferSize:   bufferSize,
		fromTailFlag: fromTailFlag,
		cursor:       &Cursor{},
	}

	inode, device, err := reader.getStatInfo()

	if err != nil {
		return nil, err
	}

	reader.setCursor(
		&Cursor{
			Inode:  inode,
			Device: device,
		},
	)

	// check whether we should use storage cursor
	cursorValidFlag := inode == initialCursor.Inode && device == initialCursor.Device

	if cursorValidFlag {
		reader.setCursor(initialCursor)
	}

	err = reader.acquireSource(!cursorValidFlag)

	if err != nil {
		return nil, err
	}

	return &reader, nil
}

// EntryRead returns one line from file, cutting off line ending if it's not the last line and tries to handle rotation
func (reader *LineReader) EntryRead() ([]byte, bool, error) {
	if !reader.GetAcquireFlag() {
		return nil, false, fmt.Errorf("journal from specified path '%s' isn't acquired", reader.filePath)
	}

	buffer, prefixFlag, err := reader.buffer.ReadLine()

	if int64(len(buffer)) != 0 {
		position, errSeek := reader.fileHandler.Seek(0, 1)

		// should not happen normally
		if errSeek != nil {
			return nil, false, errSeek
		}

		reader.setCursor(
			&Cursor{
				Inode:  reader.cursor.Inode,
				Device: reader.cursor.Device,
				Value:  position - int64(reader.buffer.Buffered()),
			},
		)
		result := make([]byte, len(buffer))
		copy(result, buffer)

		return result, prefixFlag, nil
	}

	if err != io.EOF {
		return nil, false, err
	}

	// check if file has been rotated or deleted
	inode, device, err := reader.getStatInfo()

	if err != nil {
		reader.setCursor(&Cursor{})
		return nil, false, err
	}

	if inode == reader.cursor.Inode && device == reader.cursor.Device {
		return nil, false, nil
	}

	reader.setCursor(
		&Cursor{
			Inode:  inode,
			Device: device,
			Value:  0,
		},
	)
	err = reader.acquireSource(false)

	if err != nil {
		return nil, false, err
	}

	return nil, false, nil
}

func (reader *LineReader) acquireSource(considerTailFlag bool) error {
	reader.Close()
	fileHandler, err := os.Open(reader.filePath)

	if err != nil {
		return err
	}

	if considerTailFlag && reader.fromTailFlag {
		reader.cursor.Value = GetLastSeparatorPosition(fileHandler)
	}

	offset, err := fileHandler.Seek(reader.cursor.Value, 0)

	if err != nil {
		return err
	}

	reader.cursor.Value = offset
	buffer := bufio.NewReaderSize(fileHandler, reader.bufferSize)
	reader.fileHandler = fileHandler
	reader.buffer = buffer
	return nil
}

func (reader *LineReader) getStatInfo() (uint64, uint64, error) {
	stat, err := os.Stat(reader.filePath)

	if err != nil {
		return 0, 0, err
	}

	statInfo, ok := stat.Sys().(*syscall.Stat_t)

	if !ok {
		return 0, 0, err
	}

	return statInfo.Ino, statInfo.Dev, nil
}

func (reader *LineReader) setCursor(cursor *Cursor) {
	reader.cursor = cursor
}

// GetCursor should be used to get current reader in-memory cursor
func (reader *LineReader) GetCursor() *Cursor {
	return reader.cursor
}

// GetAcquireFlag should be used to check if ReaderJournald is able to provide data
func (reader *LineReader) GetAcquireFlag() bool {
	return reader.fileHandler != nil
}

// Close should be used as LineReader finalizer
func (reader *LineReader) Close() error {
	var err error

	if reader.fileHandler != nil {
		err = reader.fileHandler.Close()
	}

	reader.fileHandler = nil
	reader.buffer = nil
	return err
}

// GetLastSeparatorPosition returns first occurrence of '\n', reading file chunk by chunk from the end
// returns zero in case of IO errors (unlikely) or missing separator
// note: does not keep current file cursor for easiness’ sake
func GetLastSeparatorPosition(fileHandler io.ReadSeeker) int64 {
	offset, _ := fileHandler.Seek(0, 2)
	step := int64(16)

	for offset > 0 {
		// file's origin is reached
		if offset < step {
			step = offset
		}

		offset -= step
		_, err := fileHandler.Seek(offset, 0)

		if err != nil {
			return 0
		}

		buffer := make([]byte, step)
		_, err = fileHandler.Read(buffer)

		if err != nil {
			return 0
		}

		position := bytes.LastIndexByte(buffer, byte(10))

		if position < 0 {
			continue
		}

		return offset + int64(position+1)
	}

	return 0
}
