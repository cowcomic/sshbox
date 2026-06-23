package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestConnectionJSON(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	conn := Connection{
		Name:      "test-server",
		Host:      "192.168.1.1",
		Port:      22,
		User:      "root",
		Password:  "encrypted",
		Tags:      []string{"prod", "web"},
		Notes:     "test server",
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(conn)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded Connection
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.Name != conn.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, conn.Name)
	}
	if decoded.Host != conn.Host {
		t.Errorf("Host = %q, want %q", decoded.Host, conn.Host)
	}
	if decoded.Port != conn.Port {
		t.Errorf("Port = %d, want %d", decoded.Port, conn.Port)
	}
	if len(decoded.Tags) != len(conn.Tags) {
		t.Errorf("Tags len = %d, want %d", len(decoded.Tags), len(conn.Tags))
	}
}

func TestConfigJSON(t *testing.T) {
	cfg := Config{
		Version: "1.0.0",
		Connections: []Connection{
			{Name: "s1", Host: "1.1.1.1", Port: 22, User: "root"},
		},
		LastUpdated: time.Now().Truncate(time.Second),
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded Config
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", decoded.Version, "1.0.0")
	}
	if len(decoded.Connections) != 1 {
		t.Errorf("Connections len = %d, want 1", len(decoded.Connections))
	}
}
