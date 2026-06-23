package storage

import (
	"os"
	"path/filepath"
	"runtime"
	"sshbox/internal/models"
	"testing"
)

func tempStore(t *testing.T) *ConfigStore {
	t.Helper()
	dir := t.TempDir()
	return NewConfigStoreWithPath(filepath.Join(dir, "config.json"))
}

func TestLoadNonExistent(t *testing.T) {
	store := tempStore(t)
	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Version != configVersion {
		t.Errorf("Version = %q, want %q", cfg.Version, configVersion)
	}
	if len(cfg.Connections) != 0 {
		t.Errorf("Connections len = %d, want 0", len(cfg.Connections))
	}
}

func TestSaveAndLoad(t *testing.T) {
	store := tempStore(t)
	cfg := &models.Config{
		Version: configVersion,
		Connections: []models.Connection{
			{Name: "test", Host: "1.1.1.1", Port: 22, User: "root"},
		},
	}

	if err := store.Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if len(loaded.Connections) != 1 {
		t.Fatalf("Connections len = %d, want 1", len(loaded.Connections))
	}
	if loaded.Connections[0].Name != "test" {
		t.Errorf("Name = %q, want %q", loaded.Connections[0].Name, "test")
	}
}

func TestFilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permissions not enforced on Windows")
	}

	store := tempStore(t)
	cfg := &models.Config{Version: configVersion, Connections: []models.Connection{}}
	if err := store.Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	info, err := os.Stat(store.Path())
	if err != nil {
		t.Fatalf("Stat() error: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("file permissions = %o, want 0600", perm)
	}
}

func TestBackup(t *testing.T) {
	store := tempStore(t)
	cfg := &models.Config{Version: configVersion, Connections: []models.Connection{}}
	if err := store.Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	if err := store.Backup(); err != nil {
		t.Fatalf("Backup() error: %v", err)
	}

	// Check backup file exists
	entries, err := os.ReadDir(filepath.Dir(store.Path()))
	if err != nil {
		t.Fatalf("ReadDir() error: %v", err)
	}

	found := false
	for _, e := range entries {
		if e.Name() != "config.json" && !e.IsDir() {
			found = true
			break
		}
	}
	if !found {
		t.Error("backup file not found")
	}
}

func TestBackupNonExistent(t *testing.T) {
	store := tempStore(t)
	// Backup of non-existent file should be no-op
	if err := store.Backup(); err != nil {
		t.Fatalf("Backup() error: %v", err)
	}
}
