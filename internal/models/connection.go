package models

import "time"

// Connection represents an SSH connection configuration.
type Connection struct {
	Name      string    `json:"name"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	User      string    `json:"user"`
	Password  string    `json:"password"` // encrypted
	Tags      []string  `json:"tags"`
	Notes     string    `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Config represents the sshbox configuration.
type Config struct {
	Version     string       `json:"version"`
	Connections []Connection `json:"connections"`
	LastUpdated time.Time    `json:"last_updated"`
}
