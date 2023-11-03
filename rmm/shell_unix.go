//go:build !windows
// +build !windows

package rmm

func getShellCommand() []string {
	return []string{"bash"}
}
