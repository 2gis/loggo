package k8s

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/2gis/loggo/configuration"
	"github.com/2gis/loggo/logging"
)

func TestNew(t *testing.T) {
	client := fake.NewSimpleClientset()
	c := configuration.SLIExporterConfig{ServiceDefaultDomain: "2gis.test"}

	provider := NewProviderK8SServices(client, c, logging.NewLoggerDefault())

	assert.IsType(t, &ProviderK8SServices{}, provider)
}

func TestRetrieveServices(t *testing.T) {
	client := getFakeKubernetesClient(false, false)
	provider := NewProviderK8SServices(client, defaultAnnotationsConfig(), logging.NewLoggerDefault())
	provider.Retrieve()

	assert.Len(t, provider.services, 3)
	fmt.Print(provider.services)
	assert.Equal(t, "test", provider.services["test.2gis.ru"].Name)
	assert.Equal(t, "test.2gis.ru", provider.services["test.2gis.ru"].Domains[0])
	assert.Len(t, provider.services["test.2gis.ru"].Domains, 2)
	assert.Equal(t, true, provider.services["test.2gis.ru"].Enabled)
	assert.Equal(t, "metrics", provider.services["test.2gis.ru"].Paths[0].Label)
}

func TestRetrieveServicesWithEmptyAnnotations(t *testing.T) {
	client := getFakeKubernetesClient(false, true)
	provider := NewProviderK8SServices(client, defaultAnnotationsConfig(), logging.NewLoggerDefault())
	provider.Retrieve()

	assert.Len(t, provider.services, 0)
}

func TestFailingOfCreatingService(t *testing.T) {
	client := getFakeKubernetesClient(true, false)
	provider := NewProviderK8SServices(client, defaultAnnotationsConfig(), logging.NewLoggerDefault())
	provider.Retrieve()

	assert.Len(t, provider.services, 0)
}

func TestGetServiceByHost(t *testing.T) {
	client := getFakeKubernetesClient(false, false)
	provider := NewProviderK8SServices(client, defaultAnnotationsConfig(), logging.NewLoggerDefault())
	provider.Retrieve()

	s := provider.GetServiceByHost("test.2gis.ru")

	assert.Len(t, provider.services, 3)
	assert.Equal(t, "test", s.Name)
	assert.Equal(t, "test.2gis.ru", s.Domains[0])
	assert.Len(t, s.Domains, 2)
	assert.Equal(t, true, s.Enabled)
	assert.Equal(t, "metrics", s.Paths[0].Label)

	s = provider.GetServiceByHost("test.2gis.test")

	assert.Len(t, provider.services, 3)
	assert.Equal(t, "test", s.Name)
	assert.Equal(t, "test.2gis.ru", s.Domains[0])
	assert.Len(t, s.Domains, 2)
	assert.Equal(t, true, s.Enabled)
	assert.Equal(t, "metrics", s.Paths[0].Label)

	assert.Nil(t, provider.GetServiceByHost("not-exist"))
}

func getFakeKubernetesClient(broken bool, empty bool) *fake.Clientset {
	annotations := make(map[string]string)
	if !empty {
		annotations["loggo.sla/enable"] = "enable"
		if broken {
			annotations["loggo.sla/paths"] = "/metrics"
		} else {
			annotations["loggo.sla/paths"] = `[{"metrics":[".*"]}]`
		}
		annotations["router.deis.io/domains"] = "test.2gis.ru,test"
	}
	services := &core.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:        "test",
			Namespace:   "io",
			Annotations: annotations},
	}

	namespaces := &core.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: "io",
		},
	}
	return fake.NewSimpleClientset(services, namespaces)
}

func defaultAnnotationsConfig() configuration.SLIExporterConfig {
	return configuration.SLIExporterConfig{
		AnnotationExporterPaths:  configuration.AnnotationExporterPathsDefault,
		AnnotationExporterEnable: configuration.AnnotationExporterEnableDefault,
		AnnotationSLADomains:     configuration.AnnotationSLADomainsDefault,
		ServiceDefaultDomain:     "2gis.test",
	}

}
