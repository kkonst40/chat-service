package util

import (
	"os"
	"path/filepath"
)

func GetCurrentDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	exeDir := filepath.Dir(exePath)
	return exeDir, nil
}
