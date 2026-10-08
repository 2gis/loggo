package containers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/logging"
)

// Matched files are expected to be located at .../<namespace>_<pod>_<id>/<container>/<n>.log

// Container represents container configuration
type Container struct {
	Type string

	ID      string
	LogPath string
	State   StateSection
	Config  ConfigSection
}

// Running returns value of corresponding field of the container config
func (c *Container) Running() bool {
	return c.State.Running
}

// Containers is the map of container entities
type Containers map[string]*Container

// Present checks whether the key present in containers map
func (containers Containers) Present(path string) bool {
	_, ok := containers[path]
	return ok
}

// ConfigSection represents Config section of container configuration
type ConfigSection struct {
	Labels map[string]string
}

// StateSection represents State section of container configuration
type StateSection struct {
	Running bool
}

// ProviderContainers seeks for logs matching include glob patterns and not matching exclude ones
type ProviderContainers struct {
	includes []string
	excludes []string
	logger   logging.Logger
}

// GetPodName returns container pod name or empty string
func (c *Container) GetPodName() string {
	return c.getLabelValue(common.LabelKubernetesPodName)
}

// GetPodNamespace returns container pod namespace or empty string
func (c *Container) GetPodNamespace() string {
	return c.getLabelValue(common.LabelKubernetesPodNamespace)
}

// GetName returns container name or empty string
func (c *Container) GetName() string {
	return c.getLabelValue(common.LabelKubernetesContainerName)
}

func (c *Container) getLabelValue(label string) string {
	if value, ok := c.Config.Labels[label]; ok {
		return value
	}
	return ""
}

// NewProviderContainers is ProviderContainers constructor.
// Both arguments are comma-separated lists of glob patterns (see filepath.Match);
// '*' does not match path separators.
func NewProviderContainers(include, exclude string, logger logging.Logger) (*ProviderContainers, error) {
	includes := splitPatterns(include)
	excludes := splitPatterns(exclude)

	if len(includes) == 0 {
		return nil, fmt.Errorf("logs path must contain at least one pattern")
	}

	for _, pattern := range append(append([]string{}, includes...), excludes...) {
		if _, err := filepath.Match(pattern, ""); err != nil {
			return nil, fmt.Errorf("invalid pattern '%s': %w", pattern, err)
		}
	}

	logger.Infof("Searching logs by patterns: %v, excluding: %v", includes, excludes)

	return &ProviderContainers{
		includes: includes,
		excludes: excludes,
		logger:   logger,
	}, nil
}

func splitPatterns(value string) []string {
	patterns := make([]string, 0)

	for _, pattern := range strings.Split(value, ",") {
		if pattern = strings.TrimSpace(pattern); pattern != "" {
			patterns = append(patterns, pattern)
		}
	}

	return patterns
}

// Containers seek and return all Containers
func (provider *ProviderContainers) Containers() (Containers, error) {
	containers := make(Containers)

	for _, pattern := range provider.includes {
		paths, err := filepath.Glob(pattern)

		if err != nil {
			return containers, err
		}

		for _, path := range paths {
			if provider.excluded(path) || !isRegularFile(path) {
				continue
			}

			container, ok := deserializeContainerConfigContainerD(path)
			if !ok {
				provider.logger.Warnf("containers provider: unexpected log path layout, skipping: %s", path)
				continue
			}

			containers[container.LogPath] = container
		}
	}

	return containers, nil
}

func (provider *ProviderContainers) excluded(path string) bool {
	for _, pattern := range provider.excludes {
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
	}

	return false
}

// isRegularFile reports whether path is a non-symlink, non-directory file
func isRegularFile(path string) bool {
	info, err := os.Lstat(path)

	return err == nil && info.Mode().IsRegular()
}

func deserializeContainerConfigContainerD(path string) (*Container, bool) {
	containerDir := filepath.Dir(path)

	split := strings.Split(filepath.Base(filepath.Dir(containerDir)), "_")
	if len(split) != 3 {
		return nil, false
	}

	namespace := split[0]
	pod := split[1]
	id := split[2]

	return &Container{
		Type:    common.CRITypeContainerD,
		ID:      id,
		LogPath: path,
		Config: ConfigSection{
			Labels: map[string]string{
				common.LabelKubernetesPodName:       pod,
				common.LabelKubernetesPodNamespace:  namespace,
				common.LabelKubernetesContainerName: filepath.Base(containerDir),
			},
		},
		// todo: hotfix, until we'll be able to evaluate it
		State: StateSection{Running: true},
	}, true
}
