//go:build windows
// +build windows

package system

import "fmt"

func unsupportedOnWindows(operation string) error {
	return fmt.Errorf("%s is not supported on Windows; the plugin relies on Unix-style mkfs, mount, and filesystem tooling", operation)
}

func GetDiskUsage(path string) (int64, error) {
	return 0, unsupportedOnWindows("disk usage queries")
}

func GetFSStats(path string) (FSStats, error) {
	return FSStats{}, unsupportedOnWindows("filesystem statistics")
}

func IsMounted(path string) (bool, error) {
	return false, unsupportedOnWindows("mount checks")
}

func MountImage(path, volume, filesystem string) error {
	return unsupportedOnWindows("volume creation and mounting")
}

func UmountImage(path string) error {
	return unsupportedOnWindows("volume unmounting")
}

func Format(path, filesystem string) error {
	return unsupportedOnWindows("filesystem formatting")
}
