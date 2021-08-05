package redisclient

import "errors"

// ClientMock is a mock for redis client
type ClientMock struct {
	buffer []string
	closed bool

	broken bool
}

// NewClientMock is a constuctor for ClientMock
func NewClientMock(broken bool) *ClientMock {
	return &ClientMock{broken: broken}
}

// DeliverMessages stores data in buffer
func (m *ClientMock) DeliverMessages(data []string) error {
	if m.broken {
		return errors.New("exception (504) Reason: \"channel/connection is not open\"")
	}

	m.buffer = append(m.buffer, data...)
	return nil
}

// Close does nothing in mock
func (m *ClientMock) Close() error {
	if m.broken {
		return errors.New("exception (504) Reason: \"channel/connection is not open\"")
	}
	m.closed = true
	return nil
}

// GetBuffer return content of buffer in mock
func (m *ClientMock) GetBuffer() []string {
	return m.buffer
}
