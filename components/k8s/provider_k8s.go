package k8s

import (
	"context"
	"fmt"
	"strings"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/2gis/loggo/configuration"
	"github.com/2gis/loggo/logging"
)

// ProviderK8SServices serves to retrieve and provide data about K8S services leveraging K8S Api server through go-client
type ProviderK8SServices struct {
	sync.Mutex
	clientSet kubernetes.Interface
	services  map[string]*Service
	config    configuration.SLIExporterConfig
	logger    logging.Logger
}

// NewProviderK8SServices is a constructor for ProviderK8SServices
func NewProviderK8SServices(
	client kubernetes.Interface, config configuration.SLIExporterConfig, logger logging.Logger) *ProviderK8SServices {
	return &ProviderK8SServices{
		clientSet: client,
		config:    config,
		logger:    logger,
	}
}

// Retrieve retrieves objects from k8s and refreshes inner mapping
func (p *ProviderK8SServices) Retrieve() error {
	p.Lock()
	defer p.Unlock()

	p.services = make(map[string]*Service)
	namespaces, err := p.clientSet.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		p.logger.Debugf("Error: Unable to list namespaces: %s", err.Error())
		return err
	}

	for _, namespace := range namespaces.Items {
		services, err := p.clientSet.CoreV1().Services(namespace.GetName()).List(context.TODO(),metav1.ListOptions{})

		if err != nil {
			p.logger.Debugf("Unable to list service in namespace '%s': %s",
				namespace.GetName(),
				err.Error(),
			)
			return err
		}

		p.logger.Debugf(
			"Process services from namespace '%s': %d",
			namespace.GetName(),
			len(services.Items),
		)

		for _, item := range services.Items {
			service, err := CreateService(p.config, item.GetObjectMeta().GetAnnotations())

			if err != nil {
				p.logger.Warnf(
					"Unable to use service '%s.%s', %s",
					item.GetObjectMeta().GetNamespace(),
					item.GetObjectMeta().GetName(),
					err.Error(),
				)
				continue
			}

			if service == nil {
				continue
			}

			service.Name = item.GetObjectMeta().GetName()

			for _, domain := range service.Domains {
				if !strings.Contains(domain, ".") {
					p.services[fmt.Sprintf("%s.%s", domain, p.config.ServiceDefaultDomain)] = service
				}

				p.services[domain] = service
			}
		}
	}
	p.logger.Debugf("Services in registry: %v", p.services)
	return nil
}

// GetServiceByHost returns service from inner map by host, if any
func (p *ProviderK8SServices) GetServiceByHost(host string) *Service {
	p.Lock()
	defer p.Unlock()

	service, ok := p.services[host]

	if !ok {
		return nil
	}

	return service
}
