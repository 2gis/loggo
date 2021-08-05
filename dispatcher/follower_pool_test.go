package dispatcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/2gis/loggo/dispatcher/workers"
	"github.com/2gis/loggo/tests/mocks"
)

func TestFollowerPool(t *testing.T) {
	path0 := "path_0"
	path1 := "path_1"

	cases := map[string]workers.Follower{
		path0: &mocks.WorkerFollowerMock{},
		path1: &mocks.WorkerFollowerMock{},
	}

	fp := NewFollowerPool()

	for key, value := range cases {
		fp.Add(key, value)
		assert.Contains(t, fp.Pool(), key)

		follower, ok := fp.Get(key)
		assert.True(t, ok)
		assert.Equal(t, follower, value)
	}

	fp.Remove(path0)
	_, ok := fp.Get(path0)
	assert.False(t, ok)
	assert.Len(t, fp.Pool(), 1)
}
