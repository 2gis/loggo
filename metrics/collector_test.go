package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBucketsSplit(t *testing.T) {
	bucketsSlice, err := buckets(".01 0.03 0.05 0.07 0.1 0.15 0.2 0.25 0.3 0.4 0.5 0.6 7")
	assert.NoError(t, err)
	assert.Equal(t, []float64{0.01, 0.03, 0.05, 0.07, 0.1, 0.15, 0.2, 0.25, 0.3, 0.4, 0.5, 0.6, 7}, bucketsSlice)

	_, err = buckets(",.01 0.03 0.05 0.07 0.1 0.15 0.2 0.25 0.3 0.4 0.5 0.6 7")
	assert.Error(t, err)
}
