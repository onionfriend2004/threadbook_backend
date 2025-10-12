package config

import (
	"mime"
	"path/filepath"
	"strings"
)

type FileConfig struct {
	Common struct {
		AllowedFormats []string
		MaxSizeBytes   int64
	}
	Spool struct {
		MaxBannerSizeBytes int64
	}
}

func NewFileConfig(cfg *Config) *FileConfig {
	return &FileConfig{
		Common: struct {
			AllowedFormats []string
			MaxSizeBytes   int64
		}{
			AllowedFormats: cfg.Upload.Common.AllowedFormats,
			MaxSizeBytes:   int64(cfg.Upload.Common.MaxSizeMB) << 20,
		},
		Spool: struct {
			MaxBannerSizeBytes int64
		}{
			MaxBannerSizeBytes: int64(cfg.Upload.Spool.MaxBannerSizeMB) << 20,
		},
	}
}

func (f *FileConfig) IsAllowedFormat(filename string) bool {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	for _, format := range f.Common.AllowedFormats {
		if strings.ToLower(format) == ext {
			return true
		}
	}
	return false
}

func (f *FileConfig) GetContentTypeByExtension(filename string) string {
	ext := filepath.Ext(filename)
	return mime.TypeByExtension(ext)
}

func (f *FileConfig) ValidateSize(fileType string, size int64) bool {
	switch fileType {
	case "spool_banner":
		return size <= f.Spool.MaxBannerSizeBytes
	default:
		return size <= f.Common.MaxSizeBytes
	}
}

func (f *FileConfig) GetMaxSize(fileType string) int64 {
	switch fileType {
	case "spool_banner":
		return f.Spool.MaxBannerSizeBytes
	default:
		return f.Common.MaxSizeBytes
	}
}

func (f *FileConfig) GetAllowedFormats() []string {
	return f.Common.AllowedFormats
}
