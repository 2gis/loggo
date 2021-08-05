package rates

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// RateRecordsProvider is the implementation of records provider that tries to get them from yaml file
type RateRecordsProviderYaml struct {
	filePath string
}

// NewRateRecordsProviderYaml is the constructor for RateRecordsProvider
func NewRateRecordsProviderYaml(filePath string) *RateRecordsProviderYaml {
	return &RateRecordsProviderYaml{
		filePath: filePath,
	}
}

// RateRecords should be used to obtain records from list containing yaml
func (provider *RateRecordsProviderYaml) RateRecords() ([]RateRecord, error) {
	yamlData, err := ioutil.ReadFile(provider.filePath)

	if err != nil {
		return nil, err
	}

	result := make([]RateRecord, 0)
	err = yaml.Unmarshal(yamlData, &result)

	if err != nil {
		return nil, err
	}

	return result, nil
}
