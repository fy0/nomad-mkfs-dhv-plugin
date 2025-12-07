package config

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/viper"
)

const Version = "0.1.0"

const (
	DefaultFilesystem   = "ext4"
	DefaultBlockSize    = ""
	DefaultMountOptions = ""
	DefaultSparse       = true // 默认使用稀疏文件
)

type DynamicHostVolumeConfig struct {
	Operation string `mapstructure:"OPERATION"`

	VolumesDir string `mapstructure:"VOLUMES_DIR"`
	VolumeID   string `mapstructure:"VOLUME_ID"`
	PluginDir  string `mapstructure:"PLUGIN_DIR"`
	Namespace  string `mapstructure:"NAMESPACE"`
	VolumeName string `mapstructure:"VOLUME_NAME"`
	NodeID     string `mapstructure:"NODE_ID"`
	NodePool   string `mapstructure:"NODE_POOL"`
	Parameters string `mapstructure:"DHV_PARAMETERS"`

	CapacityMinBytes int64 `mapstructure:"CAPACITY_MIN_BYTES"`
	CapacityMaxBytes int64 `mapstructure:"CAPACITY_MAX_BYTES"`

	CreatedPath string `mapstructure:"CREATED_PATH"`
}

type DynamicHostVolumeParameters struct {
	FileSystem   string `json:"filesystem"`
	BlockSize    string `json:"block_size"`
	MountOptions string `json:"mount_options"`
	ReadOnly     bool   `json:"read_only"`
	Sparse       bool   `json:"sparse"` // 是否使用稀疏文件
}

func SetupDynamicHostVolumeConfig() (DynamicHostVolumeConfig, error) {
	mpstruct := viper.New()
	cfg := newDefault()

	mpstruct.SetConfigType("env")
	mpstruct.SetEnvPrefix("DHV")
	mpstruct.AutomaticEnv()

	if err := mpstruct.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("unable to unmarshal config: %w", err)
	}

	return cfg, nil
}

func (cfg *DynamicHostVolumeConfig) GetParams() (*DynamicHostVolumeParameters, error) {
	// 打印接收到的原始参数
	log.Printf("========================================")
	log.Printf("DEBUG: Raw DHV_PARAMETERS received from Nomad:")
	log.Printf("  Value: '%s'", cfg.Parameters)
	log.Printf("  Length: %d bytes", len(cfg.Parameters))
	log.Printf("  Empty: %v", cfg.Parameters == "")
	log.Printf("========================================")

	params := &DynamicHostVolumeParameters{
		FileSystem:   DefaultFilesystem,
		BlockSize:    DefaultBlockSize,
		MountOptions: DefaultMountOptions,
		ReadOnly:     false,
		Sparse:       DefaultSparse, // 默认启用稀疏文件
	}

	if cfg.Parameters != "" {
		log.Printf("DEBUG: Attempting to parse parameters as JSON...")
		if err := json.Unmarshal([]byte(cfg.Parameters), &params); err != nil {
			log.Printf("DEBUG: JSON parse FAILED: %v", err)
			return nil, fmt.Errorf("unable to parse parameters as json: %w", err)
		}
		log.Printf("DEBUG: JSON parse SUCCESS")
		log.Printf("DEBUG: Parsed parameters: %+v", params)
	} else {
		log.Printf("DEBUG: No parameters provided, using defaults")
	}

	log.Printf("DEBUG: Final parameters to use:")
	log.Printf("  FileSystem:   %s", params.FileSystem)
	log.Printf("  BlockSize:    %s", params.BlockSize)
	log.Printf("  MountOptions: %s", params.MountOptions)
	log.Printf("  ReadOnly:     %v", params.ReadOnly)
	log.Printf("  Sparse:       %v", params.Sparse)
	log.Printf("========================================")

	return params, nil
}

func newDefault() DynamicHostVolumeConfig {
	return DynamicHostVolumeConfig{
		Operation: "",

		VolumesDir: "",
		VolumeID:   "",
		PluginDir:  "",
		Namespace:  "",
		VolumeName: "",
		NodeID:     "",
		NodePool:   "",

		Parameters: "{}",

		CapacityMinBytes: 0,
		CapacityMaxBytes: 0,

		CreatedPath: "",
	}
}
