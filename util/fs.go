package util

import (
	"os"
	"path/filepath"
)

func CreateParentDir(path string) error {
	parentDirPath := filepath.Dir(path)
	err := os.MkdirAll(parentDirPath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
