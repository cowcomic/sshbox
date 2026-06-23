package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sshbox/internal/models"
	"time"
)

const (
	configFileName = "config.json"
	configVersion  = "1.0.0"
	filePerm       = 0600
)

// ConfigStore manages reading and writing the sshbox configuration.
type ConfigStore struct {
	path string
}

// NewConfigStore creates a ConfigStore using the executable's directory.
func NewConfigStore() (*ConfigStore, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("获取可执行文件路径失败: %w", err)
	}
	return &ConfigStore{path: filepath.Join(filepath.Dir(exe), configFileName)}, nil
}

// NewConfigStoreWithPath creates a ConfigStore with a custom path (for testing).
func NewConfigStoreWithPath(path string) *ConfigStore {
	return &ConfigStore{path: path}
}

// Load reads the config from disk. Returns an empty config if the file doesn't exist.
func (s *ConfigStore) Load() (*models.Config, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &models.Config{
				Version:     configVersion,
				Connections: []models.Connection{},
				LastUpdated: time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg models.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	return &cfg, nil
}

// Save writes the config to disk with 600 permissions.
func (s *ConfigStore) Save(cfg *models.Config) error {
	cfg.LastUpdated = time.Now()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	if err := os.WriteFile(s.path, data, filePerm); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}
	return nil
}

// Backup creates a timestamped backup of the config file.
func (s *ConfigStore) Backup() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // nothing to backup
		}
		return fmt.Errorf("读取配置文件失败: %w", err)
	}
	backupPath := s.path + "." + time.Now().Format("20060102150405") + ".bak"
	return os.WriteFile(backupPath, data, filePerm)
}

// Path returns the config file path.
func (s *ConfigStore) Path() string {
	return s.path
}
