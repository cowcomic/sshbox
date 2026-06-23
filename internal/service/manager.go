package service

import (
	"encoding/json"
	"fmt"
	"os"
	"sshbox/internal/crypto"
	"sshbox/internal/models"
	"sshbox/internal/storage"
	"strings"
	"time"
)

// ConnectionManager manages SSH connection configurations.
type ConnectionManager struct {
	store     *storage.ConfigStore
	masterKey []byte
}

// NewConnectionManager creates a new ConnectionManager.
func NewConnectionManager(store *storage.ConfigStore, masterKey []byte) *ConnectionManager {
	return &ConnectionManager{store: store, masterKey: masterKey}
}

// AddConnection adds a new SSH connection.
func (m *ConnectionManager) AddConnection(name, host string, port int, user, password string, tags []string, notes string) error {
	cfg, err := m.store.Load()
	if err != nil {
		return err
	}

	for _, c := range cfg.Connections {
		if c.Name == name {
			return fmt.Errorf("连接 %s 已存在", name)
		}
	}

	encrypted, err := crypto.Encrypt(password, m.masterKey)
	if err != nil {
		return fmt.Errorf("加密密码失败: %w", err)
	}

	now := time.Now()
	conn := models.Connection{
		Name:      name,
		Host:      host,
		Port:      port,
		User:      user,
		Password:  encrypted,
		Tags:      tags,
		Notes:     notes,
		CreatedAt: now,
		UpdatedAt: now,
	}

	cfg.Connections = append(cfg.Connections, conn)
	return m.store.Save(cfg)
}

// ListConnections returns connections, optionally filtered by tag or search term.
func (m *ConnectionManager) ListConnections(tag, search string) ([]models.Connection, error) {
	cfg, err := m.store.Load()
	if err != nil {
		return nil, err
	}

	result := cfg.Connections
	if tag != "" {
		var filtered []models.Connection
		for _, c := range result {
			for _, t := range c.Tags {
				if t == tag {
					filtered = append(filtered, c)
					break
				}
			}
		}
		result = filtered
	}

	if search != "" {
		search = strings.ToLower(search)
		var filtered []models.Connection
		for _, c := range result {
			if strings.Contains(strings.ToLower(c.Name), search) ||
				strings.Contains(strings.ToLower(c.Host), search) ||
				strings.Contains(strings.ToLower(c.Notes), search) ||
				matchTag(c.Tags, search) {
				filtered = append(filtered, c)
			}
		}
		result = filtered
	}

	return result, nil
}

// GetConnection returns a connection by name.
func (m *ConnectionManager) GetConnection(name string) (*models.Connection, error) {
	cfg, err := m.store.Load()
	if err != nil {
		return nil, err
	}

	for _, c := range cfg.Connections {
		if c.Name == name {
			return &c, nil
		}
	}
	return nil, fmt.Errorf("连接 %s 不存在", name)
}

// UpdateConnection updates fields of an existing connection.
func (m *ConnectionManager) UpdateConnection(name string, host *string, port *int, user, password *string, tags *[]string, notes *string) error {
	cfg, err := m.store.Load()
	if err != nil {
		return err
	}

	for i, c := range cfg.Connections {
		if c.Name == name {
			if host != nil {
				cfg.Connections[i].Host = *host
			}
			if port != nil {
				cfg.Connections[i].Port = *port
			}
			if user != nil {
				cfg.Connections[i].User = *user
			}
			if password != nil {
				encrypted, err := crypto.Encrypt(*password, m.masterKey)
				if err != nil {
					return fmt.Errorf("加密密码失败: %w", err)
				}
				cfg.Connections[i].Password = encrypted
			}
			if tags != nil {
				cfg.Connections[i].Tags = *tags
			}
			if notes != nil {
				cfg.Connections[i].Notes = *notes
			}
			cfg.Connections[i].UpdatedAt = time.Now()
			return m.store.Save(cfg)
		}
	}
	return fmt.Errorf("连接 %s 不存在", name)
}

// DeleteConnection removes a connection by name.
func (m *ConnectionManager) DeleteConnection(name string) error {
	cfg, err := m.store.Load()
	if err != nil {
		return err
	}

	for i, c := range cfg.Connections {
		if c.Name == name {
			cfg.Connections = append(cfg.Connections[:i], cfg.Connections[i+1:]...)
			return m.store.Save(cfg)
		}
	}
	return fmt.Errorf("连接 %s 不存在", name)
}

// GetDecryptedPassword returns the decrypted password for a connection.
func (m *ConnectionManager) GetDecryptedPassword(name string) (string, error) {
	conn, err := m.GetConnection(name)
	if err != nil {
		return "", err
	}
	return crypto.Decrypt(conn.Password, m.masterKey)
}

// ListTags returns all tags with their connection counts.
func (m *ConnectionManager) ListTags() (map[string]int, error) {
	cfg, err := m.store.Load()
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int)
	for _, c := range cfg.Connections {
		for _, t := range c.Tags {
			counts[t]++
		}
	}
	return counts, nil
}

// ExportJSON exports connections as JSON, optionally filtered by tag.
func (m *ConnectionManager) ExportJSON(tag string) ([]byte, error) {
	conns, err := m.ListConnections(tag, "")
	if err != nil {
		return nil, err
	}

	export := struct {
		Version     string              `json:"version"`
		ExportedAt  time.Time           `json:"exported_at"`
		Connections []models.Connection `json:"connections"`
	}{
		Version:     "1.0.0",
		ExportedAt:  time.Now(),
		Connections: conns,
	}

	return json.MarshalIndent(export, "", "  ")
}

// ImportJSON imports connections from JSON data.
func (m *ConnectionManager) ImportJSON(data []byte, merge bool) (int, error) {
	var importData struct {
		Connections []models.Connection `json:"connections"`
	}
	if err := json.Unmarshal(data, &importData); err != nil {
		return 0, fmt.Errorf("解析导入文件失败: %w", err)
	}

	cfg, err := m.store.Load()
	if err != nil {
		return 0, err
	}

	if !merge {
		cfg.Connections = importData.Connections
	} else {
		existing := make(map[string]bool)
		for _, c := range cfg.Connections {
			existing[c.Name] = true
		}
		for _, c := range importData.Connections {
			if !existing[c.Name] {
				cfg.Connections = append(cfg.Connections, c)
			}
		}
	}

	return len(importData.Connections), m.store.Save(cfg)
}

// ImportFromFile reads a file and imports its contents.
func (m *ConnectionManager) ImportFromFile(path string, merge bool) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("读取导入文件失败: %w", err)
	}
	return m.ImportJSON(data, merge)
}

func matchTag(tags []string, search string) bool {
	for _, t := range tags {
		if strings.Contains(strings.ToLower(t), search) {
			return true
		}
	}
	return false
}
