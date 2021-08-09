package containers

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/2gis/loggo/logging"
)

type testEnvironment struct {
	tempDir       string
	linksDir      string
	containersDir string
	logFileDir    string
	logFileName   string
	containerDir  string
}

func setupEnvironment(t *testing.T) testEnvironment {
	environment := testEnvironment{}
	environment.tempDir = filepath.Join(os.TempDir(), "loggo-tests")
	environment.linksDir = filepath.Join(environment.tempDir, "loggo-pods")
	environment.containersDir = filepath.Join(environment.tempDir, "loggo-containers")
	err := os.MkdirAll(environment.linksDir, 0755)
	assert.NoError(t, err)
	err = os.MkdirAll(environment.containersDir, 0755)
	assert.NoError(t, err)

	id := "123abc"
	environment.logFileName = fmt.Sprintf("%s-json.log", id)
	environment.logFileDir = filepath.Join(environment.containersDir, id)
	err = os.MkdirAll(environment.logFileDir, 0755)
	assert.NoError(t, err)

	inflatedDir := filepath.Join(environment.logFileDir, "inflated")
	err = os.MkdirAll(inflatedDir, 0755)
	assert.NoError(t, err)

	environment.containerDir = filepath.Join(environment.linksDir, "95d6c1ec-323f-11e8-822e-fa163e24fbac")
	err = os.MkdirAll(environment.containerDir, 0755)
	assert.NoError(t, err)

	data := []byte(`{"log":"Hello world"}`)
	ioutil.WriteFile(filepath.Join(environment.logFileDir, environment.logFileName), data, 0700)
	ioutil.WriteFile(filepath.Join(inflatedDir, environment.logFileName), data, 0700)

	data = []byte(`{
    "ID":"123abc",
    "Config": {
      "Labels":{
        "io.kubernetes.pod.namespace":"yabloko",
        "io.kubernetes.pod.name":"123abc",
        "io.kubernetes.container.name": "service"
      }
    }
  }`)

	ioutil.WriteFile(filepath.Join(environment.logFileDir, configFileName), data, 0700)
	ioutil.WriteFile(filepath.Join(inflatedDir, configFileName), data, 0700)
	os.Symlink(
		filepath.Join(
			environment.logFileDir,
			environment.logFileName,
		),
		filepath.Join(
			environment.containerDir,
			"my-service_0.log",
		),
	)
	os.Symlink(
		filepath.Join(
			inflatedDir,
			environment.logFileName,
		),
		filepath.Join(
			environment.containerDir,
			"my-service_1.log",
		),
	)

	return environment
}

func tearDown(tempDir string) {
	os.RemoveAll(tempDir)
}

func TestFunctions(t *testing.T) {
	environment := setupEnvironment(t)
	defer tearDown(environment.tempDir)

	directories, err := Tree(environment.linksDir)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(directories))
	assert.Equal(t, environment.containerDir, directories[0])

	links, _, err := SymlinksAndFiles(directories[0])
	assert.NoError(t, err)
	assert.Equal(t, 2, len(links))
	assert.Equal(t, filepath.Join(environment.containerDir, "my-service_0.log"), links[0])
	assert.Equal(t, filepath.Join(environment.containerDir, "my-service_1.log"), links[1])

	actualPath, err := os.Readlink(links[0])
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(environment.logFileDir, environment.logFileName), actualPath)

	actualPath, err = os.Readlink(links[1])
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(filepath.Join(environment.logFileDir, "inflated"), environment.logFileName), actualPath)

	configPath, err := getConfigFilePath(actualPath)
	assert.NoError(t, err)

	config, err := deserializeContainerConfig(actualPath, configPath)
	assert.NoError(t, err)
	assert.Equal(t, actualPath, config.LogPath)
	assert.Equal(t, "123abc", config.ID)
	assert.Equal(t, "yabloko", config.GetPodNamespace())
	assert.Equal(t, "123abc", config.GetPodName())
	assert.Equal(t, "service", config.GetName())
}

func TestContainersProvider(t *testing.T) {
	te := setupEnvironment(t)
	defer tearDown(te.tempDir)

	providerContainers, err := NewProviderContainers(te.linksDir, logging.NewLoggerDefault())
	assert.NoError(t, err)

	containers, err := providerContainers.Containers()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(containers))

	fmt.Println(containers)
	container := containers["/tmp/loggo-tests/loggo-containers/123abc/123abc-json.log"]
	assert.Equal(t, "service", container.GetName())
	assert.Equal(t, "123abc", container.GetPodName())
	assert.Equal(t, "yabloko", container.GetPodNamespace())
}
