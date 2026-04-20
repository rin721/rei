package storage

import (
	"fmt"
	"os"
	"strconv"
)

// Config 保存文件服务配置
type Config struct {
	// FSType 文件系统类型 (os, memory, readonly, basepath)
	FSType FSType `mapstructure:"fs_type"`

	// BasePath 基础路径,用于 basepath 文件系统类型
	BasePath string `mapstructure:"base_path"`

	// EnableWatch 是否启用文件监听功能
	EnableWatch bool `mapstructure:"enable_watch"`

	// WatchBufferSize 文件监听事件缓冲区大小
	WatchBufferSize int `mapstructure:"watch_buffer_size"`
}

// ValidateName 返回配置名称
func (c *Config) ValidateName() string {
	return "storage"
}

// Validate 验证配置有效性
func (c *Config) Validate() error {
	// 验证文件系统类型
	switch c.FSType {
	case FSTypeOS, FSTypeMemory, FSTypeReadOnly, FSTypeBasePathFS:
		// 有效类型
	default:
		return fmt.Errorf("%w: %s", ErrInvalidFSType, c.FSType)
	}

	// 验证基础路径
	if c.FSType == FSTypeBasePathFS && c.BasePath == "" {
		return fmt.Errorf("%w: base_path is required for basepath filesystem", ErrInvalidConfig)
	}

	// 验证监听缓冲区大小
	if c.WatchBufferSize < 0 {
		return fmt.Errorf("%w: watch_buffer_size must be non-negative", ErrInvalidConfig)
	}

	return nil
}

// DefaultConfig 返回默认配置
func (c *Config) DefaultConfig() {
	c.FSType = DefaultFSType
	c.BasePath = DefaultBasePath
	c.EnableWatch = true
	c.WatchBufferSize = 100
}

// OverrideConfig 从环境变量覆盖配置
func (c *Config) OverrideConfig() {
	// STORAGE_FS_TYPE
	if fsType := os.Getenv("STORAGE_FS_TYPE"); fsType != "" {
		c.FSType = FSType(fsType)
	}

	// STORAGE_BASE_PATH
	if basePath := os.Getenv("STORAGE_BASE_PATH"); basePath != "" {
		c.BasePath = basePath
	}

	// STORAGE_ENABLE_WATCH
	if enableWatch := os.Getenv("STORAGE_ENABLE_WATCH"); enableWatch != "" {
		if val, err := strconv.ParseBool(enableWatch); err == nil {
			c.EnableWatch = val
		}
	}

	// STORAGE_WATCH_BUFFER_SIZE
	if bufferSize := os.Getenv("STORAGE_WATCH_BUFFER_SIZE"); bufferSize != "" {
		if val, err := strconv.Atoi(bufferSize); err == nil {
			c.WatchBufferSize = val
		}
	}
}
