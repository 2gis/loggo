package rates

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type ratesProviderMock struct {
	records   []RateRecord
	errorFlag bool
}

func newRatesProviderMock(records []RateRecord) *ratesProviderMock {
	return &ratesProviderMock{records: records}
}

func (rpm *ratesProviderMock) RateRecords() ([]RateRecord, error) {
	if rpm.errorFlag {
		return []RateRecord(nil), errors.New("error")
	}

	return rpm.records, nil
}

func TestNewRater(t *testing.T) {
	provider := NewRuleRecordsProviderStub()
	_, err := NewRater(provider, 0)
	assert.Error(t, err)

	rater, err := NewRater(provider, 1)
	assert.NoError(t, err)
	assert.Equal(
		t,
		&Rater{
			rateDefault:        1,
			provider:           provider,
			NamespaceRules:     []*NamespaceRule(nil),
			PodRules:           []*PodRule(nil),
			NamespacedPodRules: []*NamespacedPodRule(nil),
		},
		rater,
	)

}

func TestRater_RetrieveRates(t *testing.T) {
	rateDefault := 400.0
	rate := 2.0
	namespace := "namespace"
	pod := "pod"

	provider := newRatesProviderMock(
		[]RateRecord{{Namespace: namespace, Pod: pod, Rate: rate}},
	)
	rater, err := NewRater(provider, 400)
	assert.NoError(t, err)
	err = rater.Retrieve()
	assert.NoError(t, err)
	assert.Equal(t, rateDefault, rater.rateDefault)
	assert.Equal(t, rate, rater.Rate("namespace", "pod"))
	assert.Equal(t, rateDefault, rater.Rate("0", "pod"))
}
