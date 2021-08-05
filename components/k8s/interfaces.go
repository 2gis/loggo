package k8s

// ServicesProvider is an interface that allows to get services for SLA metering
type ServicesProvider interface {
	Retrieve() error
	GetServiceByHost(host string) *Service
}
