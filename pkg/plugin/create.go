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

func Create(cfg config.DynamicHostVolumeConfig) error {
	// 打印所有接收到的配置
	log.Printf("========================================")
	log.Printf("DEBUG: Create() called with configuration:")
	log.Printf("  Operation:        %s", cfg.Operation)
	log.Printf("  VolumesDir:       %s", cfg.VolumesDir)
	log.Printf("  VolumeID:         %s", cfg.VolumeID)
	log.Printf("  PluginDir:        %s", cfg.PluginDir)
	log.Printf("  Namespace:        %s", cfg.Namespace)
	log.Printf("  VolumeName:       %s", cfg.VolumeName)
	log.Printf("  NodeID:           %s", cfg.NodeID)
	log.Printf("  NodePool:         %s", cfg.NodePool)
	log.Printf("  CapacityMinBytes: %d (%d MB)", cfg.CapacityMinBytes, cfg.CapacityMinBytes/(1024*1024))
	log.Printf("  CapacityMaxBytes: %d (%d MB)", cfg.CapacityMaxBytes, cfg.CapacityMaxBytes/(1024*1024))
	log.Printf("  Parameters (raw): '%s'", cfg.Parameters)
	log.Printf("  CreatedPath:      %s", cfg.CreatedPath)
	log.Printf("========================================")

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

	volumePath := filepath.Join(cfg.VolumesDir, cfg.VolumeID)
	imagePath := fmt.Sprintf("%s.img", volumePath)

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

		// 只有在非稀疏模式下才写入初始数据
		if !params.Sparse {
			log.Printf("Writing initial data (non-sparse mode)")
			zeros := make([]byte, 1024*1024) // 1MB
			if _, err := file.Write(zeros); err != nil {
				return fmt.Errorf("failed to initialize '%s': %w", imagePath, err)
			}
		} else {
			log.Printf("Sparse file mode enabled, skipping initial write")
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
