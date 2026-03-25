//go:build !windows
// +build !windows

package system

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func GetMountDetails(path string) (MountDetails, error) {
	mounts, err := os.Open("/proc/mounts")
	if err != nil {
		return MountDetails{}, fmt.Errorf("failed to access '/proc/mounts': %w", err)
	}
	defer mounts.Close()

	scanner := bufio.NewScanner(mounts)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			continue
		}

		if unescapeMountField(fields[1]) != path {
			continue
		}

		return MountDetails{
			Source: unescapeMountField(fields[0]),
			FSType: fields[2],
		}, nil
	}

	if err := scanner.Err(); err != nil {
		return MountDetails{}, fmt.Errorf("failed to read '/proc/mounts': %w", err)
	}

	return MountDetails{}, fmt.Errorf("mount details for '%s' not found", path)
}

func ProbeFilesystemType(path string) (string, error) {
	blkid, err := FindPath("blkid")
	if err != nil {
		return "", fmt.Errorf("blkid tool not found: %w", err)
	}

	cmd := exec.Command(blkid, "-o", "value", "-s", "TYPE", path)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to probe filesystem type for '%s': %w", path, err)
	}

	return strings.TrimSpace(string(output)), nil
}

func ResizeExt4(path, imagePath, mountSource string, mounted bool) error {
	if mounted {
		target := strings.TrimSpace(mountSource)
		if target == "" {
			return fmt.Errorf("unable to determine mounted source for '%s'", path)
		}

		if strings.HasPrefix(target, "/dev/loop") {
			if err := refreshLoopDevice(target); err != nil {
				return err
			}
		}

		return runResize2fs(target)
	}

	return runResize2fs(imagePath)
}

func refreshLoopDevice(loopDevice string) error {
	losetup, err := FindPath("losetup")
	if err != nil {
		return fmt.Errorf("losetup tool not found: %w", err)
	}

	cmd := exec.Command(losetup, "-c", loopDevice)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to refresh loop device '%s': %w", loopDevice, err)
	}

	return nil
}

func runResize2fs(target string) error {
	resize2fs, err := FindPath("resize2fs")
	if err != nil {
		return fmt.Errorf("resize2fs tool not found: %w", err)
	}

	cmd := exec.Command(resize2fs, target)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run resize2fs on '%s': %w", target, err)
	}

	return nil
}

func unescapeMountField(value string) string {
	replacer := strings.NewReplacer(
		`\040`, " ",
		`\011`, "	",
		`\012`, "\n",
		`\134`, "\\",
	)

	return replacer.Replace(value)
}
