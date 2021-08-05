package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeChannelsString(t *testing.T) {
	inputA := make(chan string, 1)
	inputB := make(chan string, 1)
	output := MergeChannelsString(inputA, inputB)

	inputA <- "a"
	assert.Equal(t, "a", <-output)

	inputB <- "b"
	assert.Equal(t, "b", <-output)

	close(inputA)
	close(inputB)

	_, ok := <-output
	assert.False(t, ok)
}
