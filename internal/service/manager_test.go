package service

import (
	"path/filepath"
	"sshbox/internal/crypto"
	"sshbox/internal/storage"
	"testing"
)

var testMasterKey = []byte("0123456789abcdef0123456789abcdef")

func tempManager(t *testing.T) *ConnectionManager {
	t.Helper()
	dir := t.TempDir()
	store := storage.NewConfigStoreWithPath(filepath.Join(dir, "config.json"))
	return NewConnectionManager(store, testMasterKey)
}

func TestAddAndGetConnection(t *testing.T) {
	m := tempManager(t)

	if err := m.AddConnection("s1", "1.1.1.1", 22, "root", "pass123", []string{"prod"}, "test"); err != nil {
		t.Fatalf("AddConnection() error: %v", err)
	}

	conn, err := m.GetConnection("s1")
	if err != nil {
		t.Fatalf("GetConnection() error: %v", err)
	}

	if conn.Name != "s1" {
		t.Errorf("Name = %q, want %q", conn.Name, "s1")
	}
	if conn.Host != "1.1.1.1" {
		t.Errorf("Host = %q, want %q", conn.Host, "1.1.1.1")
	}
	if conn.Port != 22 {
		t.Errorf("Port = %d, want 22", conn.Port)
	}
	if conn.User != "root" {
		t.Errorf("User = %q, want %q", conn.User, "root")
	}
	if len(conn.Tags) != 1 || conn.Tags[0] != "prod" {
		t.Errorf("Tags = %v, want [prod]", conn.Tags)
	}
	if conn.Notes != "test" {
		t.Errorf("Notes = %q, want %q", conn.Notes, "test")
	}
}

func TestAddDuplicate(t *testing.T) {
	m := tempManager(t)
	m.AddConnection("s1", "1.1.1.1", 22, "root", "pass", nil, "")

	err := m.AddConnection("s1", "2.2.2.2", 22, "root", "pass", nil, "")
	if err == nil {
		t.Error("adding duplicate name should fail")
	}
}

func TestGetNonExistent(t *testing.T) {
	m := tempManager(t)
	_, err := m.GetConnection("nope")
	if err == nil {
		t.Error("GetConnection for non-existent should fail")
	}
}

func TestListConnections(t *testing.T) {
	m := tempManager(t)
	m.AddConnection("s1", "1.1.1.1", 22, "root", "pass", []string{"prod"}, "")
	m.AddConnection("s2", "2.2.2.2", 22, "root", "pass", []string{"dev"}, "")
	m.AddConnection("s3", "3.3.3.3", 22, "root", "pass", []string{"prod", "web"}, "")

	// List all
	all, err := m.ListConnections("", "")
	if err != nil {
		t.Fatalf("ListConnections() error: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("len = %d, want 3", len(all))
	}

	// Filter by tag
	prod, _ := m.ListConnections("prod", "")
	if len(prod) != 2 {
		t.Errorf("prod filter len = %d, want 2", len(prod))
	}

	// Search
	web, _ := m.ListConnections("", "web")
	if len(web) != 1 {
		t.Errorf("search 'web' len = %d, want 1", len(web))
	}
}

func TestUpdateConnection(t *testing.T) {
	m := tempManager(t)
	m.AddConnection("s1", "1.1.1.1", 22, "root", "pass", []string{"old"}, "old notes")

	newHost := "2.2.2.2"
	newPort := 2222
	newUser := "admin"
	newPass := "newpass"
	newTags := []string{"new"}
	newNotes := "new notes"

	if err := m.UpdateConnection("s1", &newHost, &newPort, &newUser, &newPass, &newTags, &newNotes); err != nil {
		t.Fatalf("UpdateConnection() error: %v", err)
	}

	conn, _ := m.GetConnection("s1")
	if conn.Host != "2.2.2.2" {
		t.Errorf("Host = %q, want %q", conn.Host, "2.2.2.2")
	}
	if conn.Port != 2222 {
		t.Errorf("Port = %d, want 2222", conn.Port)
	}
	if conn.User != "admin" {
		t.Errorf("User = %q, want %q", conn.User, "admin")
	}
	if conn.Notes != "new notes" {
		t.Errorf("Notes = %q, want %q", conn.Notes, "new notes")
	}
}

func TestUpdateNonExistent(t *testing.T) {
	m := tempManager(t)
	h := "1.1.1.1"
	err := m.UpdateConnection("nope", &h, nil, nil, nil, nil, nil)
	if err == nil {
		t.Error("UpdateConnection for non-existent should fail")
	}
}

func TestDeleteConnection(t *testing.T) {
	m := tempManager(t)
	m.AddConnection("s1", "1.1.1.1", 22, "root", "pass", nil, "")

	if err := m.DeleteConnection("s1"); err != nil {
		t.Fatalf("DeleteConnection() error: %v", err)
	}

	_, err := m.GetConnection("s1")
	if err == nil {
		t.Error("GetConnection after delete should fail")
	}
}

func TestDeleteNonExistent(t *testing.T) {
	m := tempManager(t)
	err := m.DeleteConnection("nope")
	if err == nil {
		t.Error("DeleteConnection for non-existent should fail")
	}
}

func TestGetDecryptedPassword(t *testing.T) {
	m := tempManager(t)
	m.AddConnection("s1", "1.1.1.1", 22, "root", "mypassword", nil, "")

	pass, err := m.GetDecryptedPassword("s1")
	if err != nil {
		t.Fatalf("GetDecryptedPassword() error: %v", err)
	}
	if pass != "mypassword" {
		t.Errorf("password = %q, want %q", pass, "mypassword")
	}
}

func TestListTags(t *testing.T) {
	m := tempManager(t)
	m.AddConnection("s1", "1.1.1.1", 22, "root", "pass", []string{"prod", "web"}, "")
	m.AddConnection("s2", "2.2.2.2", 22, "root", "pass", []string{"prod"}, "")
	m.AddConnection("s3", "3.3.3.3", 22, "root", "pass", []string{"dev"}, "")

	tags, err := m.ListTags()
	if err != nil {
		t.Fatalf("ListTags() error: %v", err)
	}

	if tags["prod"] != 2 {
		t.Errorf("prod count = %d, want 2", tags["prod"])
	}
	if tags["web"] != 1 {
		t.Errorf("web count = %d, want 1", tags["web"])
	}
	if tags["dev"] != 1 {
		t.Errorf("dev count = %d, want 1", tags["dev"])
	}
}

func TestExportImport(t *testing.T) {
	m := tempManager(t)
	m.AddConnection("s1", "1.1.1.1", 22, "root", "pass", []string{"prod"}, "notes1")
	m.AddConnection("s2", "2.2.2.2", 22, "root", "pass", []string{"dev"}, "notes2")

	// Export all
	data, err := m.ExportJSON("")
	if err != nil {
		t.Fatalf("ExportJSON() error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("exported data is empty")
	}

	// Export by tag
	prodData, _ := m.ExportJSON("prod")
	if len(prodData) == 0 {
		t.Fatal("prod export is empty")
	}

	// Import into fresh manager
	m2 := tempManager(t)
	count, err := m2.ImportJSON(data, false)
	if err != nil {
		t.Fatalf("ImportJSON() error: %v", err)
	}
	if count != 2 {
		t.Errorf("import count = %d, want 2", count)
	}

	conns, _ := m2.ListConnections("", "")
	if len(conns) != 2 {
		t.Errorf("after import len = %d, want 2", len(conns))
	}
}

func TestImportMerge(t *testing.T) {
	m := tempManager(t)
	m.AddConnection("s1", "1.1.1.1", 22, "root", "pass", nil, "")

	// Import with merge - s1 exists, s2 is new
	importData := []byte(`{"connections":[{"name":"s1","host":"9.9.9.9","port":22,"user":"root","password":"x"},{"name":"s2","host":"2.2.2.2","port":22,"user":"root","password":"y"}]}`)
	count, err := m.ImportJSON(importData, true)
	if err != nil {
		t.Fatalf("ImportJSON(merge) error: %v", err)
	}
	if count != 2 {
		t.Errorf("import count = %d, want 2", count)
	}

	conns, _ := m.ListConnections("", "")
	if len(conns) != 2 {
		t.Errorf("after merge len = %d, want 2", len(conns))
	}

	// s1 should keep original host
	s1, _ := m.GetConnection("s1")
	if s1.Host != "1.1.1.1" {
		t.Errorf("s1 host = %q, want %q (should keep original)", s1.Host, "1.1.1.1")
	}
}

func TestImportOverwrite(t *testing.T) {
	m := tempManager(t)
	m.AddConnection("s1", "1.1.1.1", 22, "root", "pass", nil, "")

	importData := []byte(`{"connections":[{"name":"new1","host":"9.9.9.9","port":22,"user":"root","password":"x"}]}`)
	count, err := m.ImportJSON(importData, false)
	if err != nil {
		t.Fatalf("ImportJSON(overwrite) error: %v", err)
	}
	if count != 1 {
		t.Errorf("import count = %d, want 1", count)
	}

	conns, _ := m.ListConnections("", "")
	if len(conns) != 1 {
		t.Errorf("after overwrite len = %d, want 1", len(conns))
	}
	if conns[0].Name != "new1" {
		t.Errorf("name = %q, want %q", conns[0].Name, "new1")
	}
}

func TestImportInvalidJSON(t *testing.T) {
	m := tempManager(t)
	_, err := m.ImportJSON([]byte("not json"), false)
	if err == nil {
		t.Error("ImportJSON with invalid JSON should fail")
	}
}

// Ensure crypto package is used (for encryption verification)
func TestPasswordEncrypted(t *testing.T) {
	m := tempManager(t)
	m.AddConnection("s1", "1.1.1.1", 22, "root", "plaintext", nil, "")

	// Read raw config to verify password is encrypted
	store := m.store
	cfg, _ := store.Load()
	conn := cfg.Connections[0]

	// Password should not be plaintext
	if conn.Password == "plaintext" {
		t.Error("password should be encrypted, not plaintext")
	}

	// But should decrypt correctly
	decrypted, err := crypto.Decrypt(conn.Password, testMasterKey)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}
	if decrypted != "plaintext" {
		t.Errorf("decrypted = %q, want %q", decrypted, "plaintext")
	}
}

// Expose store for testing
func (m *ConnectionManager) Store() *storage.ConfigStore {
	return m.store
}
