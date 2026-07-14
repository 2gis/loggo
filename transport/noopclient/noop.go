package noopclient

// NoopClient is a transport client that discards all messages.
// Use it when you want to calculate metrics (SLI) without shipping logs anywhere.
type NoopClient struct{}

func NewNoopClient() *NoopClient {
	return &NoopClient{}
}

func (c *NoopClient) DeliverMessages(_ []string) error {
	return nil
}

func (c *NoopClient) Close() error {
	return nil
}
