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

func Resize(cfg config.DynamicHostVolumeConfig) error {
	path, err := resolveVolumePath(cfg)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("volume path '%s' does not exist", path)
	} else if err != nil {
		return fmt.Errorf("failed to access volume path '%s': %w", path, err)
	}

	imagePath := path + ".img"
	imageInfo, err := os.Stat(imagePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("volume image '%s' does not exist", imagePath)
	} else if err != nil {
		return fmt.Errorf("failed to access volume image '%s': %w", imagePath, err)
	}

	targetBytes, err := resolveResizeBytes(cfg)
	if err != nil {
		return err
	}

	currentBytes := imageInfo.Size()
	if targetBytes < currentBytes {
		return fmt.Errorf("shrinking is not supported: current image size is %d bytes, requested %d bytes", currentBytes, targetBytes)
	}

	mounted, err := system.IsMounted(path)
	if err != nil {
		return fmt.Errorf("failed to check mount for '%s': %w", path, err)
	}

	params, err := cfg.GetParams()
	if err != nil {
		log.Printf("Warning: Unable to parse parameters, using defaults: %v", err)
	}

	filesystem := config.DefaultFilesystem
	if params != nil && params.FileSystem != "" {
		filesystem = params.FileSystem
	}

	var mountSource string
	if mounted {
		mountDetails, err := system.GetMountDetails(path)
		if err != nil {
			return fmt.Errorf("failed to inspect mount details for '%s': %w", path, err)
		}
		mountSource = mountDetails.Source
		if mountDetails.FSType != "" {
			if filesystem != mountDetails.FSType {
				log.Printf("Warning: Configured filesystem '%s' differs from mounted filesystem '%s'; using detected filesystem", filesystem, mountDetails.FSType)
			}
			filesystem = mountDetails.FSType
		}
	} else {
		probedFilesystem, err := system.ProbeFilesystemType(imagePath)
		if err != nil {
			log.Printf("Warning: Unable to probe filesystem for '%s': %v", imagePath, err)
		} else if probedFilesystem != "" {
			if filesystem != probedFilesystem {
				log.Printf("Warning: Configured filesystem '%s' differs from probed filesystem '%s'; using detected filesystem", filesystem, probedFilesystem)
			}
			filesystem = probedFilesystem
		}
	}

	if filesystem != "ext4" {
		return fmt.Errorf("resize currently supports only ext4 filesystems, got '%s'", filesystem)
	}

	if targetBytes == currentBytes {
		return printResizeResponse(VolumeResizeResponse{
			Path:       path,
			OldBytes:   currentBytes,
			NewBytes:   targetBytes,
			FileSystem: filesystem,
			Mounted:    mounted,
		})
	}

	log.Printf("Growing volume image '%s' from %d to %d bytes", imagePath, currentBytes, targetBytes)
	if err := os.Truncate(imagePath, targetBytes); err != nil {
		return fmt.Errorf("failed to resize image '%s': %w", imagePath, err)
	}

	if err := system.ResizeExt4(path, imagePath, mountSource, mounted); err != nil {
		return fmt.Errorf("failed to resize ext4 filesystem for '%s': %w", path, err)
	}

	return printResizeResponse(VolumeResizeResponse{
		Path:       path,
		OldBytes:   currentBytes,
		NewBytes:   targetBytes,
		FileSystem: filesystem,
		Mounted:    mounted,
	})
}

func resolveVolumePath(cfg config.DynamicHostVolumeConfig) (string, error) {
	if cfg.CreatedPath != "" {
		return cfg.CreatedPath, nil
	}

	if cfg.VolumesDir == "" {
		return "", fmt.Errorf("variable 'DHV_VOLUMES_DIR' must not be empty when 'DHV_CREATED_PATH' is not provided")
	}

	if cfg.VolumeID == "" {
		return "", fmt.Errorf("variable 'DHV_VOLUME_ID' must not be empty when 'DHV_CREATED_PATH' is not provided")
	}

	return filepath.Join(cfg.VolumesDir, cfg.VolumeID), nil
}

func resolveResizeBytes(cfg config.DynamicHostVolumeConfig) (int64, error) {
	switch {
	case cfg.ResizeBytes > 0:
		return cfg.ResizeBytes, nil
	case cfg.CapacityMaxBytes > 0:
		return cfg.CapacityMaxBytes, nil
	case cfg.CapacityMinBytes > 0:
		return cfg.CapacityMinBytes, nil
	default:
		return 0, fmt.Errorf("one of 'DHV_RESIZE_BYTES', 'DHV_CAPACITY_MAX_BYTES', or 'DHV_CAPACITY_MIN_BYTES' must be greater than zero")
	}
}

func printResizeResponse(response VolumeResizeResponse) error {
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	fmt.Print(string(jsonResponse))
	return nil
}
