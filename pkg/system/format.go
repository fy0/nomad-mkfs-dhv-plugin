package system

import (
	"fmt"
	"os"
	"os/exec"
)

func Format(path, filesystem string) error {
	lookup, err := FindPath("mkfs." + filesystem)
	if err != nil {
		return fmt.Errorf("formatting tool for '%s' not found: %w", filesystem, err)
	}

	var cmd *exec.Cmd
	if filesystem == "ext4" {
		// 默认创建无日志的 ext4 文件系统
		cmd = exec.Command(lookup, "-O", "^has_journal", path)
	} else {
		cmd = exec.Command(lookup, path)
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to format '%s' with filesystem '%s': %w", path, filesystem, err)
	}

	return nil
}
