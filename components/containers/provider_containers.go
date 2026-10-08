package containers

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/logging"
)

// Container logs are expected at <logsPath>/<namespace>_<pod>_<id>/<container>/<n>.log

const (
	loggoContainerName = "loggo"
	logFilesSuffix     = ".log"
)

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

// ProviderContainers seeks for logs in requested logPath and resolves links
type ProviderContainers struct {
	logsPath string
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

// NewProviderContainers is ProviderContainers constructor
func NewProviderContainers(path string, logger logging.Logger) (*ProviderContainers, error) {
	absPath, err := filepath.Abs(path)

	if err != nil {
		return nil, err
	}

	logger.Infof("Absolute path to search logs in: '%s'", absPath)

	return &ProviderContainers{
		logsPath: absPath,
		logger:   logger,
	}, nil
}

// Containers seek and return all Containers
func (provider *ProviderContainers) Containers() (Containers, error) {
	containers := make(Containers)
	directories, err := Tree(provider.logsPath)

	if err != nil {
		return containers, err
	}

	for _, dir := range directories {
		files, err := Files(dir)

		if err != nil {
			provider.logger.Warnf("containers provider is unable to read dir: %s", dir)
			continue
		}

		for _, path := range files {
			container := deserializeContainerConfigContainerD(path)
			if !strings.HasSuffix(container.LogPath, logFilesSuffix) {
				continue
			}

			if strings.Contains(container.GetName(), loggoContainerName) {
				continue
			}

			containers[container.LogPath] = container
		}
	}

	return containers, nil
}

func Tree(path string) ([]string, error) {
	directories := make([]string, 0, 1)
	files, err := ioutil.ReadDir(path)

	if err != nil {
		return directories, err
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		filePath := filepath.Join(path, file.Name())
		subdirectories, err := Tree(filePath)

		if err != nil {
			return directories, err
		}

		directories = append(directories, filePath)
		directories = append(directories, subdirectories...)
	}

	return directories, nil
}

// Files returns regular (non-symlink) files located directly in the given directory
func Files(path string) ([]string, error) {
	files := make([]string, 0)

	content, err := ioutil.ReadDir(path)

	if err != nil {
		return files, err
	}

	for _, file := range content {
		if file.IsDir() || file.Mode()&os.ModeSymlink != 0 {
			continue
		}

		files = append(files, filepath.Join(path, file.Name()))
	}

	return files, nil
}

func deserializeContainerConfigContainerD(path string) *Container {
	containerDir := filepath.Dir(path)

	split := strings.Split(filepath.Base(filepath.Dir(containerDir)), "_")
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
	}
}
