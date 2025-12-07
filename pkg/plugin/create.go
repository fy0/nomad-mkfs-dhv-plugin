package plugin

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mwantia/nomad-mkfs-dhv-plugin/pkg/config"
	"github.com/mwantia/nomad-mkfs-dhv-plugin/pkg/system"
)

func Create(cfg config.DynamicHostVolumeConfig) error {
	if cfg.VolumesDir == "" {
		return fmt.Errorf("variable 'DHV_VOLUMES_DIR' must not be empty")
	}
	if cfg.VolumeID == "" {
		return fmt.Errorf("variable 'DHV_VOLUME_ID' must not be empty")
	}

	if cfg.CapacityMinBytes <= 0 {
		return fmt.Errorf("variable 'DHV_CAPACITY_MIN_BYTES' must be greater than zero")
	}
	if cfg.CapacityMinBytes > cfg.CapacityMaxBytes {
		return fmt.Errorf("variable 'DHV_CAPACITY_MIN_BYTES' can not be greater than 'DHV_CAPACITY_MAX_BYTES'")
	}

	params, err := cfg.GetParams()
	if err != nil {
		log.Printf("Warning: Unable to parse parameters, using defaults: %v", err)
	}

	// 生成镜像文件名: yyyymmdd-shortuuid-volumename.img
	dateStr := time.Now().Format("20060102")
	shortID := cfg.VolumeID[:8]
	imageName := fmt.Sprintf("%s-%s-%s.img", dateStr, shortID, cfg.VolumeName)

	volumePath := filepath.Join(cfg.VolumesDir, cfg.VolumeID)
	imagePath := filepath.Join(cfg.VolumesDir, imageName)

	if err := os.MkdirAll(volumePath, 0o755); err != nil {
		return fmt.Errorf("failed to create volume directory: %v", err)
	}

	capacityMB := cfg.CapacityMinBytes / (1024 * 1024)
	if capacityMB <= 0 {
		capacityMB = 1 // Ensure at least 1MB
	}

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		log.Printf("Creating filesystem image at %s (sparse: %v)", imagePath, params.Sparse)

		file, err := os.Create(imagePath)
		if err != nil {
			return fmt.Errorf("failed to create '%s': %w", imagePath, err)
		}
		defer file.Close()

		if err := file.Truncate(cfg.CapacityMinBytes); err != nil {
			return fmt.Errorf("failed to set size for '%s': %w", imagePath, err)
		}

		zeros := make([]byte, 1024*1024) // 1MB
		if _, err := file.Write(zeros); err != nil {
			return fmt.Errorf("failed to initialize '%s': %w", imagePath, err)
		}

		if err := system.Format(imagePath, params.FileSystem); err != nil {
			if err := os.Remove(imagePath); err != nil && !os.IsNotExist(err) {
				log.Printf("Warning: Failed to perform cleanup: %v", err)
			}

			return fmt.Errorf("failed to format '%s' to '%s': %w", imagePath, params.FileSystem, err)
		}
	} else {
		log.Printf("Using existing filesystem image at '%s'", imagePath)
	}

	mounted, err := system.IsMounted(volumePath)
	if err != nil {
		return fmt.Errorf("failed to check mount for '%s': %w", volumePath, err)
	}

	if !mounted {
		if err := system.MountImage(imagePath, volumePath, params.FileSystem); err != nil {
			return fmt.Errorf("failed to mount volume '%s': %w", volumePath, err)
		}
	}

	log.Printf("Mounted '%s' at '%s'", imagePath, volumePath)

	fileInfo, err := os.Stat(imagePath)
	if err != nil {
		return fmt.Errorf("failed to get filesystem size: %w", err)
	}
	actualBytes := fileInfo.Size()

	response := VolumeCreateResponse{
		Path:  volumePath,
		Bytes: actualBytes,
	}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	fmt.Print(string(jsonResponse))
	return nil
}
