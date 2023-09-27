package config

import (
	"os"
	"path/filepath"
)

var subdir = "default"

func SetSubdir(s string) {
	subdir = s
}

func GetSubdir() string {
	return subdir
}

func GetConfigDir() string {
	if os.Getenv("OS") == "Windows_NT" {
		return filepath.Join(os.Getenv("APPDATA"), "rahnit-rmm", GetSubdir())
	}
	return filepath.Join("/etc/rahnit-rmm", GetSubdir())
}

func GetFilePath(filePath ...string) string {
	pathParts := make([]string, 1, len(filePath)+1)
	pathParts[0] = GetConfigDir()
	pathParts = append(pathParts, filePath...)
	fullPath := filepath.Join(pathParts...)
	return fullPath
}
