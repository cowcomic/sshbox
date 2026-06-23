package keyring

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	gk "github.com/zalando/go-keyring"
)

const (
	serviceName = "sshbox"
	keyName     = "master-key"
)

// GetOrCreate retrieves the master key from the system keyring, or creates one if missing.
func GetOrCreate() ([]byte, error) {
	key, err := gk.Get(serviceName, keyName)
	if err == nil {
		return hex.DecodeString(key)
	}

	// Generate new 32-byte key
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, fmt.Errorf("生成主密钥失败: %w", err)
	}

	encoded := hex.EncodeToString(raw)
	if err := gk.Set(serviceName, keyName, encoded); err != nil {
		return nil, fmt.Errorf("保存主密钥到keyring失败: %w", err)
	}

	return raw, nil
}
