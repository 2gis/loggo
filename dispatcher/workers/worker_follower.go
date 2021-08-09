package workers

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/logging"
)

// workerFollower is an entity which collects log entries obtained by reader and sends them to Out channel
type workerFollower struct {
	worker

	filePath      string
	format        string
	namespace     string
	podName       string
	containerName string

	activeFlag      bool
	EOFShutdownFlag bool

	cursorStorage    Storage
	reader           LineReader
	metricsCollector MetricsCollector
	extends          common.EntryMap

	tickerCursorCommit  *time.Ticker
	tickerLimiterUpdate *time.Ticker

	sleepNoRecords time.Duration

	rater           Rater
	readRateCurrent float64
	limiter         *rate.Limiter

	output chan<- *common.Entry
	stop   chan struct{}
}

// NewFollower is a workerFollower constructor
func newFollower(output chan<- *common.Entry, filePath, format string,
	reader LineReader, collector MetricsCollector, storage Storage, rater Rater, extends common.EntryMap,
	sleepNoRecordsIntervalSec, commitIntervalSec, limitUpdateInterval int, logger logging.Logger) *workerFollower {
	podName := extends.PodName()
	containerName := extends.ContainerName()
	namespace := extends.NamespaceName()

	worker := &workerFollower{
		worker: worker{
			wg:     &sync.WaitGroup{},
			logger: logger,
		},

		filePath:      filePath,
		format:        format,
		namespace:     namespace,
		podName:       podName,
		containerName: containerName,

		cursorStorage:    storage,
		reader:           reader,
		metricsCollector: collector,
		rater:            rater,
		extends:          extends,

		sleepNoRecords:      time.Duration(sleepNoRecordsIntervalSec) * time.Second,
		tickerCursorCommit:  time.NewTicker(time.Duration(commitIntervalSec) * time.Second),
		tickerLimiterUpdate: time.NewTicker(time.Duration(limitUpdateInterval) * time.Second),

		output:     output,
		stop:       make(chan struct{}),
		activeFlag: true,
	}
	worker.setRate(rater.Rate(namespace, podName))
	return worker
}

// Start starts Follower
func (worker *workerFollower) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	worker.wg.Add(3)

	go func() {
		defer worker.wg.Done()
		worker.startCursorCommitter(ctx)
	}()
	go func() {
		defer worker.wg.Done()
		worker.startReader(ctx)
	}()
	go func() {
		defer worker.wg.Done()
		worker.startLimitUpdater(ctx)
	}()
	go worker.waitStop(cancel)

	worker.wg.Wait()
	worker.finalize()
}

func (worker *workerFollower) waitStop(cancel context.CancelFunc) {
	<-worker.stop
	cancel()
}

// Stop finishes Follower loop and lets it finalize itself; worker may be finalized by closing parent context as well
func (worker *workerFollower) Stop() {
	select {
	case worker.stop <- struct{}{}:
	default:
	}
}

// GetActiveFlag may be used to check if Follower had been finalized
func (worker *workerFollower) GetActiveFlag() bool {
	return worker.activeFlag
}

// SetEOFShutdownFlag may be used to signal worker to shutdown after first EOF
func (worker *workerFollower) SetEOFShutdownFlag() {
	worker.EOFShutdownFlag = true
}

func (worker *workerFollower) startReader(ctx context.Context) {
	worker.logger.Infof("worker on '%s' is started with rate %v mps", worker.filePath, worker.readRateCurrent)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		reservation := worker.limiter.Reserve()
		worker.metricsCollector.IncrementThrottlingDelay(
			worker.namespace,
			worker.podName,
			worker.containerName,
			float64(reservation.Delay())/float64(time.Second),
		)

		if reservation.Delay() != time.Duration(0) {
			time.Sleep(reservation.Delay())
		}

		worker.Lock()
		err := worker.entryProceed()
		worker.Unlock()

		if err != nil {
			switch err {
			case ErrNoRecords:
				if worker.EOFShutdownFlag {
					worker.Stop()
					return
				}

				time.Sleep(worker.sleepNoRecords)

			default:
				worker.logger.WithError(err).Infof(
					"reader for '%s' can't acquire file or encountered error, removing cursor, stopping worker",
					worker.filePath,
				)

				if err = worker.cursorStorage.Delete(worker.filePath); err != nil {
					worker.logger.Error(err)
				}

				worker.Stop()
				return
			}
		}
	}
}

func (worker *workerFollower) entryProceed() error {
	entry, prefixFlag, err := worker.reader.EntryRead()

	if err != nil {
		return err
	}

	if entry == nil {
		return ErrNoRecords
	}

	if prefixFlag {
		worker.logger.Warnf(
			"Reader for '%s' has encountered the long string "+
				"that doesn't fit into the buffer. Follower extends: '%s', string prefix: '%s'",
			worker.filePath,
			worker.extends,
			entry,
		)
	}

	worker.output <- &common.Entry{Origin: entry, Format: worker.format, Extends: worker.extends}
	worker.metricsCollector.IncrementLogMessageCount(
		worker.namespace,
		worker.podName,
		worker.containerName,
	)
	return nil
}

func (worker *workerFollower) startCursorCommitter(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-worker.tickerCursorCommit.C:
			worker.commitCursor()
		}
	}
}

func (worker *workerFollower) startLimitUpdater(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-worker.tickerLimiterUpdate.C:
			// i.golovchenko: must be rearranged to publisher/observers scheme later
			worker.setRate(worker.rater.Rate(worker.namespace, worker.podName))
		}
	}
}

func (worker *workerFollower) commitCursor() {
	worker.Lock()
	if err := worker.cursorStorage.Set(worker.filePath, worker.reader.GetCursor().String()); err != nil {
		worker.logger.Info(err)
	}
	worker.Unlock()
}

// setRate sets new limiter for follower
func (worker *workerFollower) setRate(readRate float64) {
	if worker.readRateCurrent == readRate {
		return
	}

	worker.readRateCurrent = readRate
	// i.golovchenko: https://github.com/golang/go/issues/23575
	worker.limiter = rate.NewLimiter(rate.Limit(readRate), int(readRate))
}

func (worker *workerFollower) finalize() {
	if worker.reader.GetAcquireFlag() {
		worker.commitCursor()
	}

	if err := worker.reader.Close(); err != nil {
		worker.logger.Warnf("worker on '%s' failed closing its reader", worker.filePath)
	}

	if !worker.metricsCollector.DeleteThrottlingDelay(
		worker.namespace,
		worker.podName,
		worker.containerName,
	) {
		worker.logger.Warnf("failed removing series with labels "+
			"namespace '%s' podName '%s' containerName '%s', file '%s'",
			worker.namespace,
			worker.podName,
			worker.containerName,
			worker.filePath,
		)
	}

	worker.tickerCursorCommit.Stop()
	worker.tickerLimiterUpdate.Stop()
	worker.activeFlag = false
}
