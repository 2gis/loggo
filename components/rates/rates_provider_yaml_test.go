package rates

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	FilePathTemp   = "/tmp/test_yaml_file.yaml"
	PayloadCorrect = `--- 
- namespace: namespace
  pod: pod
  rate: 1000

`
	PayloadInvalid = `--- 

  namespace: namespace
  rate: 1000
`
)

func TestNewRateRecordsProviderYaml(t *testing.T) {
	provider := NewRateRecordsProviderYaml("path")
	assert.Equal(t, &RateRecordsProviderYaml{filePath: "path"}, provider)
}

func TestRateRecordsPositive(t *testing.T) {
	createTestFile([]byte(PayloadCorrect))
	provider := NewRateRecordsProviderYaml(FilePathTemp)
	records, err := provider.RateRecords()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(records))
	assert.Equal(t, "namespace", records[0].Namespace)
	assert.Equal(t, "pod", records[0].Pod)
	assert.Equal(t, 1000.0, records[0].Rate)
	clean()
}

func TestRateRecordsWrongYaml(t *testing.T) {
	createTestFile([]byte(PayloadInvalid))
	provider := NewRateRecordsProviderYaml(FilePathTemp)
	_, err := provider.RateRecords()
	assert.Error(t, err)
	clean()
}

func createTestFile(payload []byte) {
	file, _ := os.Create(FilePathTemp)
	_, _ = file.Write(payload)
	file.Close()
}

func clean() {
	_ = os.Remove(FilePathTemp)
}
