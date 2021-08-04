package rates

import (
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestNewRule(t *testing.T) {
	_, err := NewRule(-1)
	assert.Error(t, err)

	_, err = NewRule(0)
	assert.Error(t, err)

	rule, err := NewRule(1)
	assert.NoError(t, err)
	assert.Equal(t, 1.0, rule.Rate)
}

func TestNamespaceRule(t *testing.T) {
	record := RateRecord{Namespace: `\w\d`, Pod: `\d\w`, Rate: 1}
	recordWrongRate := RateRecord{Namespace: `\w\d`, Pod: "", Rate: -1}
	recordWrongRegex := RateRecord{Namespace: `?`, Pod: "?", Rate: 1}

	rule, err := NewNamespaceRule(record)
	assert.NoError(t, err)
	assert.True(t, rule.Match("a1"))
	assert.False(t, rule.Match("1a"))

	_, err = NewNamespaceRule(recordWrongRate)
	assert.Error(t, err)

	_, err = NewNamespaceRule(recordWrongRegex)
	assert.Error(t, err)
}

func TestPodRule(t *testing.T) {
	record := RateRecord{Namespace: `\w\d`, Pod: `\d\w`, Rate: 1}
	recordWrongRate := RateRecord{Namespace: `\w\d`, Pod: "", Rate: -1}
	recordWrongRegex := RateRecord{Namespace: `?`, Pod: "?", Rate: 1}

	rule, err := NewPodRule(record)
	assert.NoError(t, err)
	assert.False(t, rule.Match("a1"))
	assert.True(t, rule.Match("1a"))

	_, err = NewPodRule(recordWrongRate)
	assert.Error(t, err)

	_, err = NewPodRule(recordWrongRegex)
	assert.Error(t, err)
}

func TestNamespacedPodRule(t *testing.T) {
	record := RateRecord{Namespace: `\w\d`, Pod: `\d\w`, Rate: 1}
	recordWrongRate := RateRecord{Namespace: `\w\d`, Pod: "", Rate: -1}
	recordWrongRegex := RateRecord{Namespace: `?`, Pod: "?", Rate: 1}
	recordWrongRegexPod := RateRecord{Namespace: `\w\d`, Pod: "?", Rate: 1}
	recordWrongRegexNamespace := RateRecord{Namespace: `?`, Pod: `\w\d`, Rate: 1}

	rule, err := NewNamespacedPodRule(record)
	assert.NoError(t, err)
	assert.True(t, rule.Match("a1", "1a"))
	assert.False(t, rule.Match("1a", "1a"))
	assert.False(t, rule.Match("a1", "a1"))

	_, err = NewNamespacedPodRule(recordWrongRate)
	assert.Error(t, err)
	_, err = NewNamespacedPodRule(recordWrongRegex)
	assert.Error(t, err)
	_, err = NewNamespacedPodRule(recordWrongRegexPod)
	assert.Error(t, err)
	_, err = NewNamespacedPodRule(recordWrongRegexNamespace)
	assert.Error(t, err)
}
