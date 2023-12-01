package config

import (
	"path/filepath"

	"github.com/rahn-it/svalin/util"
)

func getConfigDir() string {
	return filepath.Join(util.GetConfigDir(), "svalin")
}
