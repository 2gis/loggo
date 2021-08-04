package workers

import (
	"context"

	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/logging"
	"github.com/2gis/loggo/readers"
)

// Rater is a throttling rater interface for workers
type Rater interface {
	Rate(namespace string, pod string) float64
}

// MetricsCollector is a metrics counter object interface for workers
type MetricsCollector interface {
	IncrementLogMessageCount(namespace, podName, containerName string)
	IncrementThrottlingDelay(namespace, podName, containerName string, value float64)
	DeleteThrottlingDelay(namespace, podName, containerName string) bool
}

// Storage is the cursor storage interface for dispatcher
type Storage interface {
	Get(key string) (string, error)
	Set(key string, value string) error
	Delete(string) error
	Keys() ([]string, error)
}

// JournaldReader is specific journal reader interface
type JournaldReader interface {
	EntryRead() (entryMap common.EntryMapString, err error)
	GetAcquireFlag() bool
	GetCursor() string
	Close() error
}

// LineReader is reader interface for log follower
type LineReader interface {
	EntryRead() (entry []byte, prefixFlag bool, err error)
	GetCursor() *readers.Cursor
	GetAcquireFlag() bool
	Close() error
}

// Follower is a worker-follower interface for dispatcher
type Follower interface {
	Start(ctx context.Context)
	Stop()
	GetActiveFlag() bool
	SetEOFShutdownFlag()
}

// Follower is a worker-follower interface for dispatcher
type FollowerJournald interface {
	Start(ctx context.Context)
}

// FollowerFabric is a follower fabric interface for dispatcher
type FollowerFabric interface {
	NewFollower(
		output chan<- *common.Entry, filePath string, extends common.EntryMap) (Follower, error)
	NewFollowerJournald(
		output chan<- string, logger logging.Logger) (FollowerJournald, error)
}
