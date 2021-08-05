package dispatcher

import (
	"sync"

	"github.com/2gis/loggo/dispatcher/workers"
)

// FollowerPool is a map of Follower instances to control over
type FollowerPool struct {
	sync.RWMutex
	pool map[string]workers.Follower
}

// FollowerPool constructor
func NewFollowerPool() FollowerPool {
	return FollowerPool{pool: make(map[string]workers.Follower)}
}

// Add adds follower to pool by path
func (fp *FollowerPool) Add(path string, follower workers.Follower) {
	fp.Lock()
	defer fp.Unlock()
	fp.pool[path] = follower
}

// Remove removes follower from pool by path
func (fp *FollowerPool) Remove(path string) {
	fp.Lock()
	defer fp.Unlock()
	delete(fp.pool, path)
}

// Get returns follower and present flag by path
func (fp *FollowerPool) Get(path string) (follower workers.Follower, ok bool) {
	fp.RLock()
	defer fp.RUnlock()
	follower, ok = fp.pool[path]
	return follower, ok
}

// Pool returns copy of current pool
func (fp *FollowerPool) Pool() (pool map[string]workers.Follower) {
	fp.RLock()
	defer fp.RUnlock()

	pool = make(map[string]workers.Follower)

	for key, value := range fp.pool {
		pool[key] = value
	}

	return
}
