package k8s

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/2gis/loggo/configuration"
)

// Service is a structured "service" representation, containing service paths and domains for SLA metrics gathering
type Service struct {
	Name    string
	Enabled bool
	Paths   []PathSet
	Domains []string
}

// PathSet is shorthand structure that maps group of regular expressions with label
type PathSet struct {
	Label   string
	Regexps []*regexp.Regexp
}

// GetLabelByPath matches path to paths defined by regex group and returns regex group label
func (s *Service) GetLabelByPath(path string) string {
	for _, p := range s.Paths {
		for _, regex := range p.Regexps {
			if !regex.MatchString(path) {
				continue
			}

			return p.Label
		}
	}

	return ""
}

// CreateService tries to construct Service object from annotations dictionary
func CreateService(config configuration.SLIExporterConfig, annotations map[string]string) (*Service, error) {
	enabled, ok := annotations[config.AnnotationExporterEnable]

	if !ok {
		return nil, nil
	}

	// this parameter looks like redundant;
	// probably we can rely on checking router domains list and deprecate it in future
	if !(enabled == configuration.AnnotationExporterEnableTrue ||
		enabled == configuration.AnnotationExporterEnableTrueDeprecated) {
		return nil, nil
	}

	domains := ""

	for _, annotation := range []string{config.AnnotationSLADomains, configuration.AnnotationSLADomainsDeprecated} {
		value, ok := annotations[annotation]

		if ok {
			domains = value
			break
		}
	}

	if domains == "" {
		return nil, fmt.Errorf("there must be annotations '%s' or '%s' (deprecated) with correct domain names",
			config.AnnotationSLADomains,
			configuration.AnnotationSLADomainsDeprecated,
		)
	}

	domainsSplitted := splitDomains(domains)

	if len(domainsSplitted) == 0 {
		return nil, fmt.Errorf(
			"domains annotation is set to '%s', but the correct domain names hasn't been extracted",
			domains,
		)
	}

	if _, ok := annotations[config.AnnotationExporterPaths]; !ok {
		return nil, fmt.Errorf(
			"there must be annotations '%s' with correct syntax",
			config.AnnotationExporterPaths,
		)
	}

	paths, err := createPaths(config.AnnotationExporterPaths, annotations[config.AnnotationExporterPaths])

	if err != nil {
		return nil, err
	}

	return &Service{
		Enabled: true,
		Domains: domainsSplitted,
		Paths:   paths,
	}, nil
}

func createPaths(pathAnnotation, paths string) ([]PathSet, error) {
	var httpPaths []PathSet
	var data []map[string][]string
	paths = strings.TrimSpace(paths)

	if paths == "" {
		return httpPaths, fmt.Errorf("%s annotation is empty string", pathAnnotation)
	}

	if err := json.Unmarshal([]byte(paths), &data); err != nil {
		return httpPaths, fmt.Errorf(
			"unable to parse annotation '%s'='%s', with error '%s'",
			pathAnnotation, paths, err.Error())
	}

	for i := range data {
		pathSet := PathSet{}

		for label, regexs := range data[i] {
			if label == "" {
				return httpPaths, fmt.Errorf("can't use empty labels in '%s'='%s'", pathAnnotation, paths)
			}

			pathSet.Label = label

			if len(regexs) == 0 {
				return httpPaths, fmt.Errorf("there must be at least one regexp for label '%s' in '%s'='%s'", label, pathAnnotation, paths)
			}

			for _, regex := range regexs {
				pathRegex := strings.TrimSpace(regex)

				if len(pathRegex) == 0 {
					return httpPaths, fmt.Errorf("regexp can not be empty in '%s'='%s'", pathAnnotation, paths)
				}

				pattern, err := regexp.Compile(pathRegex)

				if err != nil {
					return httpPaths, fmt.Errorf("unable to compile regexp '%s' in '%s'='%s'", pathRegex, pathAnnotation, paths)
				}

				pathSet.Regexps = append(pathSet.Regexps, pattern)
			}

			httpPaths = append(httpPaths, pathSet)
		}
	}

	if len(httpPaths) == 0 {
		return httpPaths, fmt.Errorf("unable to use annotation '%s'='%s' it can't be empty array", pathAnnotation, paths)
	}

	return httpPaths, nil
}

func splitDomains(domains string) []string {
	if domains == "" {
		return nil
	}

	domains = strings.TrimSpace(domains)
	domains = strings.Trim(domains, ",")
	var result []string

	for _, domain := range strings.Split(domains, ",") {
		if domain != "" {
			result = append(result, strings.TrimSpace(domain))
		}
	}

	return result
}
