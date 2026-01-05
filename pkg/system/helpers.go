package system

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func FindPath(file string) (string, error) {
	paths := []string{
		"/sbin",
		"/usr/sbin",
		"/bin",
		"/usr/bin",
		"/usr/local/bin",
		"/usr/local/sbin",
	}

	for _, path := range paths {
		full := filepath.Join(path, file)
		log.Println(full)
		if _, err := os.Stat(full); err == nil {
			if IsExecutable(full) {
				return full, nil
			}
		}
	}

	return "", fmt.Errorf("file '%s' not found", file)
}

func IsExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return !info.IsDir() && (info.Mode()&0o111 != 0)
}

// GetDiskUsage returns the disk usage in bytes for the given path using du -sb
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

	// Parse output: "1234567\t/path/to/dir"
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

// GetStatfs returns filesystem statistics using syscall.Statfs
func GetStatfs(path string) (*syscall.Statfs_t, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return nil, fmt.Errorf("failed to get filesystem stats: %w", err)
	}
	return &stat, nil
}
