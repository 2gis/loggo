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

	// Build namespace ignore list when label-based filtering is configured.
	// This costs one extra Namespaces API call but keeps Services to a single request.
	var disabledNamespaces map[string]bool
	if p.config.LabelExporterNamespaceDisable != "" {
		nsList, err := p.clientSet.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{
			LabelSelector: p.config.LabelExporterNamespaceDisable,
		})
		if err != nil {
			p.logger.Debugf("Unable to list namespaces: %s", err.Error())
			return err
		}
		disabledNamespaces = make(map[string]bool, len(nsList.Items))
		for _, ns := range nsList.Items {
			disabledNamespaces[ns.GetName()] = true
		}
		p.logger.Debugf("Disabled namespaces: %v", disabledNamespaces)
	}

	services, err := p.clientSet.CoreV1().Services(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		p.logger.Debugf("Unable to list service: %s",
			err.Error(),
		)
		return err
	}

	p.logger.Debugf(
		"Process services: %d",
		len(services.Items),
	)

	for _, item := range services.Items {
		if disabledNamespaces != nil && disabledNamespaces[item.GetNamespace()] {
			continue
		}

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
