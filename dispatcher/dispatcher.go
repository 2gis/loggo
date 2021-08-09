package dispatcher

import (
	"context"
	"sync"
	"time"

	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/components/containers"
	"github.com/2gis/loggo/dispatcher/workers"
	"github.com/2gis/loggo/logging"
)

// Dispatcher starts and stops followers according to ContainersProvider targets, multiplexes their outputs
// into output; based on config, starts system journal special follower and pipes its output separately
type Dispatcher struct {
	ticker *time.Ticker

	followerPool       FollowerPool
	ignoredPool        map[string]bool
	containersProvider ContainersProvider

	followerFabric workers.FollowerFabric
	cursorStorage  workers.Storage

	startJournald bool

	output         chan *common.Entry
	outputJournald chan string

	wg     *sync.WaitGroup
	logger logging.Logger
}

// Out is a dispatcher output channel accessor
func (d *Dispatcher) Out() <-chan *common.Entry {
	return d.output
}

// OutJournald is a dispatcher outputJournald channel accessor
func (d *Dispatcher) OutJournald() <-chan string {
	return d.outputJournald
}

// NewDispatcher is a Dispatcher constructor
func NewDispatcher(
	followerFabric workers.FollowerFabric, containersProvider ContainersProvider,
	cursorStorage workers.Storage, startJournald bool, refreshInterval time.Duration, logger logging.Logger) *Dispatcher {
	return &Dispatcher{
		ticker: time.NewTicker(refreshInterval),

		followerPool: NewFollowerPool(),
		ignoredPool:  make(map[string]bool),

		followerFabric:     followerFabric,
		containersProvider: containersProvider,
		cursorStorage:      cursorStorage,
		startJournald:      startJournald,

		wg:     &sync.WaitGroup{},
		logger: logger,

		output:         make(chan *common.Entry),
		outputJournald: make(chan string),
	}
}

// Start starts Dispatcher. Dispatcher quits when context is done.
func (d *Dispatcher) Start(ctx context.Context) {
	defer close(d.output)
	defer close(d.outputJournald)

	if d.startJournald {
		d.startFollowerJournald(ctx)
	}

	d.startDispatching(ctx)
	d.wg.Wait()
	d.finalize()
}

func (d *Dispatcher) startDispatching(ctx context.Context) {
	if err := d.dispatch(ctx); err != nil {
		d.logger.Errorf("unable to get containers list, %d", err)
	}

	for {
		select {
		case <-d.ticker.C:
			if err := d.dispatch(ctx); err != nil {
				d.logger.Errorf("unable to get containers list, %d", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (d *Dispatcher) dispatch(ctx context.Context) error {
	containersActual, err := d.containersProvider.Containers()

	if err != nil {
		return err
	}

	d.removeOrphans(containersActual)
	d.startFollowers(ctx, containersActual)
	return nil
}

func (d *Dispatcher) removeOrphans(containers containers.Containers) {
	// remove orphan followers from the pool
	for path, follower := range d.followerPool.Pool() {
		if !containers.Present(path) {
			follower.Stop()
		}

		if containers.Present(path) && !containers[path].Running() {
			follower.SetEOFShutdownFlag()
			d.ignoredPool[path] = true
		}

		if !follower.GetActiveFlag() {
			d.followerPool.Remove(path)
		}
	}

	keys, err := d.cursorStorage.Keys()
	if err != nil {
		d.logger.Errorf("dispatcher: unable to read storage: %d", err)
		keys = []string{}
	}

	// remove orphan cursors from storage
	for _, key := range keys {
		// skip journald key; seems a bit ugly, but now journal key lies in the same storage
		// consider making another cursor file if possible later
		if key == common.FilenameJournald {
			continue
		}

		if containers.Present(key) {
			continue
		}

		if err := d.cursorStorage.Delete(key); err != nil {
			d.logger.Errorf("dispatcher: unable to delete key '%d' from storage: %d", key, err)
		}
	}

	// remove orphan ignore records from storage
	for path := range d.ignoredPool {
		if containers.Present(path) {
			continue
		}

		delete(d.ignoredPool, path)
	}
}

func (d *Dispatcher) startFollowers(ctx context.Context, containers containers.Containers) {
	for _, container := range containers {
		_, present := d.followerPool.Get(container.LogPath)

		if present {
			continue
		}

		if ignored := d.ignoredPool[container.LogPath]; ignored {
			continue
		}

		follower, err := d.followerFabric.NewFollower(
			d.output,
			container.LogPath,
			container.Type,
			d.containerExtends(container),
		)

		if err != nil {
			d.logger.Error(err)
			continue
		}

		d.wg.Add(1)

		go func() {
			defer d.wg.Done()
			follower.Start(ctx)
		}()

		d.followerPool.Add(container.LogPath, follower)
	}
}

func (d *Dispatcher) startFollowerJournald(ctx context.Context) {
	followerJournald, err := d.followerFabric.NewFollowerJournald(d.outputJournald, d.logger)

	if err != nil {
		d.logger.Error(err)
		return
	}

	d.wg.Add(1)

	go func() {
		defer d.wg.Done()
		followerJournald.Start(ctx)
	}()
}

func (d *Dispatcher) finalize() {
	d.ticker.Stop()
}

func (d *Dispatcher) containerExtends(c *containers.Container) (extends common.EntryMap) {
	extends = make(common.EntryMap)
	extends[common.LabelContainerID] = c.ID
	extends[common.LabelLogstashNamespace] = c.GetPodNamespace()
	extends[common.KubernetesPodName] = c.GetPodName()
	extends[common.KubernetesNamespaceName] = c.GetPodNamespace()
	extends[common.KubernetesContainerName] = c.GetName()

	return
}
