package system

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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
