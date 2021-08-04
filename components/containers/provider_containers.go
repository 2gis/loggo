package containers

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/logging"
)

// * Read all directoreis in /var/log/pods/
// * In each directory find all symlinks
// * For each symlink get destination file path
// * Read container id from destination file path
// * Get:
//    * namespace                   Config.Labels."io.kubernetes.pod.namespace"
//    * docker.container_id         ID
//    * kubernetes.pod_name         Config.Labels."io.kubernetes.pod.name"
//    * kubernetes.namespace_name   Config.Labels."io.kubernetes.pod.namespace"
//    * kubernetes.container_name   Config.Labels."io.kubernetes.container.name"

const (
	configFileName     = "config.v2.json"
	loggoContainerName = "loggo"
)

// Container represents container configuration
type Container struct {
	ID      string        `json:"ID"`
	LogPath string        `json:"LogPath"`
	State   StateSection  `json:"State"`
	Config  ConfigSection `json:"Config"`
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
	sync.Mutex
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
		links, err := Symlinks(dir)

		if err != nil {
			provider.logger.Warnf("containers provider is unable to read dir: %s", dir)
			continue
		}

		for _, link := range links {
			path, err := provider.resolveSymlink(link)

			if err != nil {
				provider.logger.Warnf("containers provider is unable to read link: %s, %s", link, err)
				continue
			}

			configPath, err := getConfigFilePath(path)

			if err != nil {
				provider.logger.Warnf("containers provider is unable get config for logfile: %s, %s", path, err)
				continue
			}

			container, err := deserializeContainerConfig(configPath)

			if err != nil {
				continue
			}

			// in current use cases regexes look like overkill
			if strings.Contains(container.GetName(), loggoContainerName) {
				continue
			}

			container.LogPath = path
			containers[container.LogPath] = container
		}
	}

	return containers, nil
}

func Tree(path string) ([]string, error) {
	output := make([]string, 0, 1)
	files, err := ioutil.ReadDir(path)

	if err != nil {
		return output, err
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		filePath := filepath.Join(path, file.Name())
		subdirectories, err := Tree(filePath)

		if err != nil {
			return output, err
		}

		output = append(output, filePath)
		output = append(output, subdirectories...)
	}

	return output, nil
}

func Symlinks(path string) ([]string, error) {
	symlinks := make([]string, 0)
	files, err := ioutil.ReadDir(path)

	if err != nil {
		return symlinks, err
	}

	for _, file := range files {
		if file.Mode()&os.ModeSymlink == 0 {
			continue
		}

		symlinks = append(symlinks, filepath.Join(path, file.Name()))
	}

	return symlinks, nil
}

func getConfigFilePath(logfile string) (string, error) {
	path, err := filepath.Abs(filepath.Dir(logfile))

	if err != nil {
		return "", err
	}

	return filepath.Join(path, configFileName), nil
}

func (provider *ProviderContainers) resolveSymlink(path string) (string, error) {
	provider.Lock()
	defer provider.Unlock()
	oldPath, err := os.Getwd()

	if err != nil {
		return "", err
	}

	newPath, err := filepath.Abs(filepath.Dir(path))

	if err != nil {
		return "", err
	}

	target, err := os.Readlink(path)

	if err != nil {
		return "", err
	}

	err = os.Chdir(newPath)
	if err != nil {
		return "", err
	}

	absTarget, err := filepath.Abs(target)

	if err != nil {
		return "", err
	}

	err = os.Chdir(oldPath)
	if err != nil {
		return "", err
	}

	return absTarget, nil
}

func deserializeContainerConfig(configPath string) (*Container, error) {
	config, err := ioutil.ReadFile(configPath)

	if err != nil {
		return nil, err
	}

	container := &Container{}
	err = json.Unmarshal(config, container)
	return container, err
}
