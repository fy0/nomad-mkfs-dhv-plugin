package plugin

const Version = "1.0.0"

type FingerprintResponse struct {
	Version string `json:"version"`
}

type VolumeCreateResponse struct {
	Path  string `json:"path"`
	Bytes int64  `json:"bytes"`
}

type VolumeUsageResponse struct {
	UsedBytes  int64 `json:"used_bytes"`
	TotalBytes int64 `json:"total_bytes"`
}
