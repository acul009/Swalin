package util

import (
	"os"
	"runtime"
)

func GetConfigDir() string {
	switch runtime.GOOS {
	case "windows":
		// use Appdata and fallback to programdata
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = os.Getenv("ProgramData")
		}
		return appData

	case "darwin":
		// also works if home is not set, since then it's just the absolute path
		return os.Getenv("HOME") + "/Library/Application Support"

	case "plan9":
		home := os.Getenv("home")
		if home != "" {
			return home + "/lib"
		}

		return "/lib"

	case "android":
		panic("not implemented")

	case "ios":
		panic("not implemented")

	default:
		xdg := os.Getenv("XDG_CONFIG_HOME")
		if xdg != "" {
			return xdg
		}

		home := os.Getenv("HOME")
		isRoot := os.Getuid() == 0
		if home == "" || isRoot {
			return "/etc"
		}

		return home + "/.local/share"
	}

}
