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

func GetMountDetails(path string) (MountDetails, error) {
	return MountDetails{}, unsupportedOnWindows("mount inspection")
}

func ProbeFilesystemType(path string) (string, error) {
	return "", unsupportedOnWindows("filesystem probing")
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

func ResizeExt4(path, imagePath, mountSource string, mounted bool) error {
	return unsupportedOnWindows("filesystem resizing")
}
