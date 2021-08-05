package transport

// Client common interface for transport
type Client interface {
	DeliverMessages(messages []string) error
	Close() error
}
