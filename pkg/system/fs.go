package system

// FSStats represents total and free bytes for the filesystem backing a path.
type FSStats struct {
	TotalBytes int64
	FreeBytes  int64
}

