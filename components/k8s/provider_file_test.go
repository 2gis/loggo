package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/2gis/loggo/configuration"
)

func TestProviderFile(t *testing.T) {
	c := configuration.SLIExporterConfig{AnnotationExporterPaths: configuration.AnnotationExporterPathsDefault,
		AnnotationExporterEnable: configuration.AnnotationExporterEnableDefault, AnnotationSLADomains: configuration.AnnotationSLADomainsDefault,
		ServiceDefaultDomain: "2gis.test", ServiceSourcePath: "./fixtures/config.yaml"}
	r, err := NewProviderFile(c)
	assert.NoError(t, err)

	empty := r.GetServiceByHost("blabla")
	assert.Nil(t, empty)

	service := r.GetServiceByHost("hello.world")
	assert.Nil(t, service)

	service = r.GetServiceByHost("hello.world1")
	assert.Nil(t, service)

	service = r.GetServiceByHost("f.2gis.test")
	assert.NotNil(t, service)
	assert.Equal(t, true, service.Enabled)
	assert.Equal(t, "f.2gis.test.ru", service.Domains[0])
	assert.Equal(t, "f.2gis.test", service.Domains[1])
	assert.Equal(t, 2, len(service.Domains))
	assert.Equal(t, "TestService1", service.Name)
	assert.Equal(t, "all", service.Paths[0].Label)
	assert.Equal(t, ".*", service.Paths[0].Regexps[0].String())

	service = r.GetServiceByHost("f.2gis.test")
	assert.Equal(t, true, service.Enabled)

	service = r.GetServiceByHost("test_service_2.2gis.test")
	assert.Equal(t, true, service.Enabled)

	service = r.GetServiceByHost("test_service_2")
	assert.Equal(t, true, service.Enabled)
}

// todo these exact asserts on error messages are insane; figure out how to get rid of them
func TestProviderFileDoesNotExist(t *testing.T) {
	c := configuration.SLIExporterConfig{ServiceDefaultDomain: "", ServiceSourcePath: "DoesNotExist"}
	_, err := NewProviderFile(c)
	assert.EqualError(t, err, "stat DoesNotExist: no such file or directory")
}

func TestRegistryFailLoadYaml(t *testing.T) {
	c := configuration.SLIExporterConfig{ServiceDefaultDomain: "", ServiceSourcePath: "./fixtures/wrongsyntax.yaml"}
	_, err := NewProviderFile(c)
	assert.EqualError(t, err, "unable to parse './fixtures/wrongsyntax.yaml': 'yaml: line 1: did not find expected node content'")

	c = configuration.SLIExporterConfig{AnnotationExporterEnable: configuration.AnnotationExporterEnableDefault, AnnotationSLADomains: configuration.AnnotationSLADomainsDefault, ServiceDefaultDomain: "2gis.test", ServiceSourcePath: "./fixtures/wrongconfig.yaml"}
	_, err = NewProviderFile(c)
	assert.EqualError(t, err, "there must be annotations 'loggo.sla/domains' or 'router.deis.io/domains' (deprecated) with correct domain names")
}
