package containers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/logging"
)

func setupEnvironment(t *testing.T) (logsDir string) {
	logsDir = t.TempDir()

	containerDir := filepath.Join(logsDir, "yabloko_123abc_95d6c1ec", "service")
	assert.NoError(t, os.MkdirAll(containerDir, 0755))
	assert.NoError(t, os.WriteFile(filepath.Join(containerDir, "0.log"), []byte("hello"), 0600))
	assert.NoError(t, os.WriteFile(filepath.Join(containerDir, "0.log.gz"), []byte("hello"), 0600))

	loggoDir := filepath.Join(logsDir, "kube-system_loggo-xyz_95d6c1ed", "loggo")
	assert.NoError(t, os.MkdirAll(loggoDir, 0755))
	assert.NoError(t, os.WriteFile(filepath.Join(loggoDir, "0.log"), []byte("hello"), 0600))

	// symlinks are not followed
	assert.NoError(t, os.Symlink(filepath.Join(containerDir, "0.log"), filepath.Join(containerDir, "1.log")))

	return logsDir
}

func TestFiles(t *testing.T) {
	logsDir := setupEnvironment(t)
	containerDir := filepath.Join(logsDir, "yabloko_123abc_95d6c1ec", "service")

	directories, err := Tree(logsDir)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(directories))

	files, err := Files(containerDir)
	assert.NoError(t, err)
	assert.Equal(t, []string{
		filepath.Join(containerDir, "0.log"),
		filepath.Join(containerDir, "0.log.gz"),
	}, files)
}

func TestContainersProvider(t *testing.T) {
	logsDir := setupEnvironment(t)

	providerContainers, err := NewProviderContainers(logsDir, logging.NewLoggerDefault())
	assert.NoError(t, err)

	containers, err := providerContainers.Containers()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(containers))

	logPath := filepath.Join(logsDir, "yabloko_123abc_95d6c1ec", "service", "0.log")
	container := containers[logPath]
	assert.Equal(t, common.CRITypeContainerD, container.Type)
	assert.Equal(t, "95d6c1ec", container.ID)
	assert.Equal(t, "service", container.GetName())
	assert.Equal(t, "123abc", container.GetPodName())
	assert.Equal(t, "yabloko", container.GetPodNamespace())
	assert.True(t, container.Running())
}
