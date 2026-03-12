//go:build !windows
// +build !windows

package system

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func IsMounted(path string) (bool, error) {
	mounts, err := os.Open("/proc/mounts")
	if err != nil {
		return false, fmt.Errorf("failed to access '/proc/mounts': %w", err)
	}
	defer mounts.Close()

	content, err := io.ReadAll(mounts)
	if err != nil {
		return false, fmt.Errorf("failed to read '/proc/mounts': %w", err)
	}

	return strings.Contains(string(content), " "+path+" "), nil
}

func MountImage(path, volume, filesystem string) error {
	mount, err := FindPath("mount")
	if err != nil {
		return fmt.Errorf("mount tool not found: %w", err)
	}

	cmd := exec.Command(mount, path, volume)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to mount '%s': %w", path, err)
	}

	return nil
}

func UmountImage(path string) error {
	if err := syscall.Unmount(path, 0); err != nil {
		if err := syscall.Unmount(path, syscall.MNT_FORCE); err != nil {
			return fmt.Errorf("failed to umount '%s': %w", path, err)
		}
	}

	return nil
}
