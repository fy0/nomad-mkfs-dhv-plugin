package config

import (
	"encoding/json"
	"fmt"

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
	params := &DynamicHostVolumeParameters{
		FileSystem:   DefaultFilesystem,
		BlockSize:    DefaultBlockSize,
		MountOptions: DefaultMountOptions,
		ReadOnly:     false,
		Sparse:       DefaultSparse,
	}

	if cfg.Parameters != "" {
		if err := json.Unmarshal([]byte(cfg.Parameters), &params); err != nil {
			return nil, fmt.Errorf("unable to parse parameters as json: %w", err)
		}
	}

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
