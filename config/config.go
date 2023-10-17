package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var subdir = "default"

var v *viper.Viper

func SetSubdir(s string) error {
	subdir = s

	_, err := os.Stat(GetConfigDir())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.MkdirAll(GetConfigDir(), 0755)
			if err != nil {
				return fmt.Errorf("failed to create config dir: %w", err)
			}
		} else {
			return fmt.Errorf("failed to check for config dir: %w", err)
		}
	}

	v = viper.New()

	_, err = os.Stat(GetFilePath("config.yml"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = createMissingConfig()
			if err != nil {
				return fmt.Errorf("failed to create config file: %w", err)
			}
		} else {
			return fmt.Errorf("failed to check for config file: %w", err)
		}
	}

	v.AddConfigPath(GetConfigDir())
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	err = v.ReadInConfig()
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	return nil
}

func Viper() *viper.Viper {
	if v == nil {
		err := fmt.Errorf("viper not initialized")
		panic(err)
	}
	return v
}

func createMissingConfig() error {
	file, err := os.Create(GetFilePath("config.yml"))
	defer file.Close()
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	return nil
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

func init() {

}
