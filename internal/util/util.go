package util

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func ReadJson(path string, dataStruct interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading config file error: %v", err)
	}

	if err := json.Unmarshal(data, dataStruct); err != nil {
		return fmt.Errorf("parsing config file error: %v", err)
	}

	return nil
}

func GetCurrentDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	exeDir := filepath.Dir(exePath)
	return exeDir, nil
}
