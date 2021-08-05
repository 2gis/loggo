package k8s

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/2gis/loggo/configuration"
)

// ProviderFile provides static map of services from file
type ProviderFile struct {
	services map[string]*Service
	config   configuration.SLIExporterConfig
}

// NewProviderFile returns new from-file provider instance
func NewProviderFile(config configuration.SLIExporterConfig) (*ProviderFile, error) {
	// file doesn't seem to be changing, so read it just once
	_, err := os.Stat(config.ServiceSourcePath)

	if err != nil {
		return nil, err
	}

	provider := &ProviderFile{config: config}
	provider.services = make(map[string]*Service)
	data, err := ioutil.ReadFile(provider.config.ServiceSourcePath)

	if err != nil {
		return nil, err
	}

	var out []map[string]string
	err = yaml.Unmarshal(data, &out)

	if err != nil {
		return nil, fmt.Errorf("unable to parse '%s': '%s'", provider.config.ServiceSourcePath, err.Error())
	}

	for _, annotations := range out {
		service, err := CreateService(config, annotations)

		if err != nil {
			return nil, err
		}

		if service == nil {
			continue
		}

		if val, ok := annotations["name"]; ok {
			service.Name = val
		}

		for _, domain := range service.Domains {
			if !strings.Contains(domain, ".") {
				provider.services[fmt.Sprintf("%s.%s", domain, provider.config.ServiceDefaultDomain)] = service
			}

			provider.services[domain] = service
		}
	}

	return provider, nil
}

// GetServiceByHost returns service from inner map by host, if any
func (provider *ProviderFile) GetServiceByHost(host string) *Service {
	service, ok := provider.services[host]

	if !ok {
		return nil
	}

	return service
}

// Retrieve does nothing in file provider
func (provider *ProviderFile) Retrieve() error {
	return nil
}
