package mocks

// CollectorMock for metrics manipulating object
type CollectorMock struct{}

func NewCollectorMock() *CollectorMock {
	return &CollectorMock{}
}

func (collector *CollectorMock) IncrementHTTPRequestCount(_, _, _, _, _ string, _ int) {}

func (collector *CollectorMock) IncrementHTTPRequestsTotalCount(_ string) {}

func (collector *CollectorMock) IncrementLogMessageCount(_, _, _ string) {}

func (collector *CollectorMock) IncrementThrottlingDelay(_, _, _ string, _ float64) {}

func (collector *CollectorMock) ObserveHTTPRequestTime(_, _, _, _, _ string, _ float64) {}

func (collector *CollectorMock) ObserveHTTPUpstreamResponseTimeTotal(_, _, _, _, _ string, _ float64) {
}

func (collector *CollectorMock) DeleteThrottlingDelay(_, _, _ string) bool {
	return true
}

func (collector *CollectorMock) Retrieve() error {
	return nil
}
