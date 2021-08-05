package k8s

// ProviderStub returns empty provider for test purposes and dry-run launches
type ProviderStub struct{}

// NewProviderStub is a constructor of ProviderStub
func NewProviderStub() *ProviderStub {
	return &ProviderStub{}
}

// Retrieve does nothing in stub
func (provider *ProviderStub) Retrieve() error {
	return nil
}

// GetServiceByHost does nothing in stub
func (provider *ProviderStub) GetServiceByHost(host string) *Service {
	return nil
}
