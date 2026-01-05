package plugin

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mwantia/nomad-mkfs-dhv-plugin/pkg/config"
	"github.com/mwantia/nomad-mkfs-dhv-plugin/pkg/system"
)

func Usage(cfg config.DynamicHostVolumeConfig) error {
	// Determine the volume path
	path := cfg.CreatedPath
	if path == "" {
		if cfg.VolumesDir == "" {
			return fmt.Errorf("variable 'DHV_VOLUMES_DIR' must not be empty when 'DHV_CREATED_PATH' is not provided")
		}
		if cfg.VolumeID == "" {
			return fmt.Errorf("variable 'DHV_VOLUME_ID' must not be empty when 'DHV_CREATED_PATH' is not provided")
		}

		path = filepath.Join(cfg.VolumesDir, cfg.VolumeID)
	}

	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("volume path '%s' does not exist", path)
	}

	// Check if it's mounted
	mounted, err := system.IsMounted(path)
	if err != nil {
		return fmt.Errorf("failed to check mount status for '%s': %w", path, err)
	}

	if !mounted {
		return fmt.Errorf("volume at '%s' is not mounted", path)
	}

	log.Printf("Getting usage for volume at %s", path)

	// Get used bytes using du
	usedBytes, err := system.GetDiskUsage(path)
	if err != nil {
		return fmt.Errorf("failed to get disk usage: %w", err)
	}

	// Get total bytes using statfs
	stat, err := system.GetStatfs(path)
	if err != nil {
		return fmt.Errorf("failed to get filesystem stats: %w", err)
	}

	// Calculate total bytes from statfs
	totalBytes := int64(stat.Blocks * uint64(stat.Bsize))

	log.Printf("Volume usage: %d/%d bytes (%.2f%%)", usedBytes, totalBytes, float64(usedBytes)/float64(totalBytes)*100)

	response := VolumeUsageResponse{
		UsedBytes:  usedBytes,
		TotalBytes: totalBytes,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	fmt.Print(string(jsonResponse))
	return nil
}
