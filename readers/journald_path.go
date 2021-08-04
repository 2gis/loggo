package readers

import (
	"os"
	"path/filepath"

	"github.com/2gis/loggo/common"
)

// JournaldPath tries to construct path to system journal from machine ID and base path
func JournaldPath(machineIDPath, journaldPath string) (string, error) {
	machineIDHandler, err := os.Open(machineIDPath)

	if err != nil {
		return "", err
	}

	machineID := make([]byte, 32)

	if _, err := machineIDHandler.Read(machineID); err != nil {
		return "", err
	}

	machineIDHandler.Close()
	return filepath.Join(journaldPath, string(machineID), common.FilenameJournald), nil
}
