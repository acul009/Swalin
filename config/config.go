package config

import (
	"fmt"
	"os"
	"path/filepath"
)

var subdir = "default"

func SetSubdir(s string) error {
	subdir = s

	err := updateViper()
	if err != nil {
		return fmt.Errorf("failed to update viper: %w", err)
	}

	return nil
}

func GetSubdir() string {
	return subdir
}

func GetConfigDir() string {
	if os.Getenv("OS") == "Windows_NT" {
		return filepath.Join(os.Getenv("APPDATA"), "github.com/rahn-it/svalin", GetSubdir())
	}
	return filepath.Join("/etc/github.com/rahn-it/svalin", GetSubdir())
}

func GetFilePath(filePath ...string) string {
	pathParts := make([]string, 1, len(filePath)+1)
	pathParts[0] = GetConfigDir()
	pathParts = append(pathParts, filePath...)
	fullPath := filepath.Join(pathParts...)
	return fullPath
}
