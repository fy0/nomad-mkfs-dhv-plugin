//go:build !windows
// +build !windows

package system

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// GetDiskUsage returns the disk usage in bytes for the given path using du -sb.
func GetDiskUsage(path string) (int64, error) {
	du, err := FindPath("du")
	if err != nil {
		return 0, fmt.Errorf("du tool not found: %w", err)
	}

	cmd := exec.Command(du, "-sb", path)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to run du: %w", err)
	}

	fields := strings.Fields(string(output))
	if len(fields) < 1 {
		return 0, fmt.Errorf("unexpected du output format")
	}

	bytes, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse du output: %w", err)
	}

	return bytes, nil
}

// GetFSStats returns filesystem total and free bytes for the given path.
func GetFSStats(path string) (FSStats, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return FSStats{}, fmt.Errorf("failed to get filesystem stats: %w", err)
	}

	blockSize := int64(stat.Bsize)
	totalBytes := int64(stat.Blocks) * blockSize
	freeBytes := int64(stat.Bfree) * blockSize

	return FSStats{TotalBytes: totalBytes, FreeBytes: freeBytes}, nil
}

