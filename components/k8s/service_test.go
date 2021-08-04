package k8s

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	c "github.com/2gis/loggo/configuration"
)

func TestGetLabelByPath(t *testing.T) {
	var paths []PathSet
	regex1 := regexp.MustCompile("abc")
	var regexs []*regexp.Regexp
	regexs = append(regexs, regex1)
	paths = append(paths, PathSet{Label: "label1", Regexps: regexs})
	s := Service{Paths: paths}
	assert.Equal(t, "label1", s.GetLabelByPath("abc"))
	assert.Equal(t, "", s.GetLabelByPath("path"))
}

func TestSplitDomains(t *testing.T) {
	output := splitDomains("a-domain.ru,b-domain.com,c-domain.ae")
	assert.Equal(t, "a-domain.ru", output[0])
	assert.Equal(t, "b-domain.com", output[1])
	assert.Equal(t, "c-domain.ae", output[2])
	output = splitDomains("")
	assert.Equal(t, 0, len(output))
	assert.Nil(t, output)
	output = splitDomains(", a-domain.ru ,")
	assert.Equal(t, "a-domain.ru", output[0])
	assert.Equal(t, 1, len(output))
	output = splitDomains(",,,,,")
	assert.Nil(t, output)
}

func TestCreatePaths(t *testing.T) {
	var emptyPaths []PathSet
	output, err := createPaths(c.AnnotationExporterPathsDefault, "")
	assert.Equal(t, emptyPaths, output)
	assert.EqualError(t, err, "loggo.sla/paths annotation is empty string")
	output, err = createPaths(c.AnnotationExporterPathsDefault, "{}")
	assert.Equal(t, emptyPaths, output)
	assert.EqualError(t, err, "unable to parse annotation 'loggo.sla/paths'='{}', with error 'json: cannot unmarshal object into Go value of type []map[string][]string'")
	output, err = createPaths(c.AnnotationExporterPathsDefault, "{")
	assert.Equal(t, emptyPaths, output)
	assert.EqualError(t, err, "unable to parse annotation 'loggo.sla/paths'='{', with error 'unexpected end of JSON input'")

	output, err = createPaths(c.AnnotationExporterPathsDefault, "[]")
	assert.Equal(t, emptyPaths, output)
	assert.EqualError(t, err, "unable to use annotation 'loggo.sla/paths'='[]' it can't be empty array")
	output, err = createPaths(c.AnnotationExporterPathsDefault, "[{\"\":[]}]")
	assert.Equal(t, emptyPaths, output)
	assert.EqualError(t, err, "can't use empty labels in 'loggo.sla/paths'='[{\"\":[]}]'")

	output, err = createPaths(c.AnnotationExporterPathsDefault, "[{\"label1\":[]}]")
	assert.Equal(t, emptyPaths, output)
	assert.EqualError(t, err, "there must be at least one regexp for label 'label1' in 'loggo.sla/paths'='[{\"label1\":[]}]'")
	output, err = createPaths(c.AnnotationExporterPathsDefault, "[{\"label1\":[\"\"]}]")
	assert.Equal(t, emptyPaths, output)
	assert.EqualError(t, err, "regexp can not be empty in 'loggo.sla/paths'='[{\"label1\":[\"\"]}]'")
	output, err = createPaths(c.AnnotationExporterPathsDefault, "[{\"label1\":[\"[\"]}]")
	assert.Equal(t, emptyPaths, output)
	assert.EqualError(t, err, "unable to compile regexp '[' in 'loggo.sla/paths'='[{\"label1\":[\"[\"]}]'")

	output, err = createPaths(c.AnnotationExporterPathsDefault, "[{\"label1\":[\".*\", \"abc\"]}, {\"label2\":[\"def\"]}]")
	assert.NoError(t, err)
	assert.Equal(t, "label1", output[0].Label)
	assert.Equal(t, "label2", output[1].Label)
	assert.Equal(t, ".*", output[0].Regexps[0].String())
	assert.Equal(t, "abc", output[0].Regexps[1].String())
	assert.Equal(t, "def", output[1].Regexps[0].String())
}

func TestCreateService(t *testing.T) {
	input := make(map[string]string)
	config := c.SLIExporterConfig{
		AnnotationExporterEnable: c.AnnotationExporterEnableDefault,
		AnnotationExporterPaths:  c.AnnotationExporterPathsDefault,
		AnnotationSLADomains:     c.AnnotationSLADomainsDefault,
	}
	output, err := CreateService(config, input)
	assert.NoError(t, err)
	assert.Nil(t, output)
	input[c.AnnotationExporterEnableDefault] = "blabla"
	output, err = CreateService(config, input)
	assert.NoError(t, err)
	assert.Nil(t, output)
	input[c.AnnotationExporterEnableDefault] = "enable"
	output, err = CreateService(config, input)
	assert.EqualError(t, err, "there must be annotations 'loggo.sla/domains' or 'router.deis.io/domains' (deprecated) with correct domain names")
	assert.Nil(t, output)
	input[c.AnnotationSLADomainsDeprecated] = "abc"
	output, err = CreateService(config, input)
	assert.EqualError(t, err, "there must be annotations 'loggo.sla/paths' with correct syntax")
	assert.Nil(t, output)

	input[c.AnnotationSLADomainsDefault] = "   "
	output, err = CreateService(config, input)
	assert.EqualError(t, err, "domains annotation is set to '   ', but the correct domain names hasn't been extracted")
	assert.Nil(t, output)

	input[c.AnnotationSLADomainsDefault] = " , "
	output, err = CreateService(config, input)
	assert.EqualError(t, err, "domains annotation is set to ' , ', but the correct domain names hasn't been extracted")
	assert.Nil(t, output)

	input[c.AnnotationSLADomainsDefault] = "abc"
	delete(input, c.AnnotationSLADomainsDeprecated)
	output, err = CreateService(config, input)
	assert.EqualError(t, err, "there must be annotations 'loggo.sla/paths' with correct syntax")
	assert.Nil(t, output)

	input[c.AnnotationExporterPathsDefault] = "[{\"label1\":[]}]"
	output, err = CreateService(config, input)
	assert.EqualError(t, err, "there must be at least one regexp for label 'label1' in 'loggo.sla/paths'='[{\"label1\":[]}]'")
	assert.Nil(t, output)

	input[c.AnnotationExporterPathsDefault] = "[{\"label1\":[\".*\", \"abc\"]}, {\"label2\":[\"def\"]}]"
	input[c.AnnotationSLADomainsDeprecated] = "abcd"
	output, err = CreateService(config, input)
	assert.NoError(t, err)
	assert.Equal(t, true, output.Enabled)
	assert.Equal(t, "abc", output.Domains[0])
	assert.Equal(t, 1, len(output.Domains))
	assert.Equal(t, "label1", output.Paths[0].Label)
	assert.Equal(t, "label2", output.Paths[1].Label)
	assert.Equal(t, ".*", output.Paths[0].Regexps[0].String())
	assert.Equal(t, "abc", output.Paths[0].Regexps[1].String())
	assert.Equal(t, "def", output.Paths[1].Regexps[0].String())
}
