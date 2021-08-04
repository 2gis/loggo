package rates

// RateRecordsProviderStub is the stub that returns empty list of rate records
type RateRecordsProviderStub struct{}

// NewRuleRecordsProviderStub is the constructor for RateRecordsProviderStub
func NewRuleRecordsProviderStub() *RateRecordsProviderStub {
	return &RateRecordsProviderStub{}
}

// RateRecords returns empty list of rate records
func (provider *RateRecordsProviderStub) RateRecords() ([]RateRecord, error) {
	return []RateRecord(nil), nil
}
