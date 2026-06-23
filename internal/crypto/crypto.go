package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	saltLen  = 16
	iterCount = 100000
	keyLen   = 32 // AES-256
)

// Encrypt encrypts plaintext using AES-256-CFB with a derived key from masterKey.
// Returns base64(salt + iv + ciphertext).
func Encrypt(plaintext string, masterKey []byte) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", fmt.Errorf("生成salt失败: %w", err)
	}

	key := pbkdf2.Key(masterKey, salt, iterCount, keyLen, sha256.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("创建cipher失败: %w", err)
	}

	pt := []byte(plaintext)
	ct := make([]byte, aes.BlockSize+len(pt))
	iv := ct[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("生成iv失败: %w", err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ct[aes.BlockSize:], pt)

	// salt + iv + ciphertext
	result := make([]byte, saltLen+aes.BlockSize+len(pt))
	copy(result, salt)
	copy(result[saltLen:], ct)

	return base64.URLEncoding.EncodeToString(result), nil
}

// Decrypt decrypts base64-encoded ciphertext using AES-256-CFB.
func Decrypt(encrypted string, masterKey []byte) (string, error) {
	data, err := base64.URLEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("base64解码失败: %w", err)
	}

	if len(data) < saltLen+aes.BlockSize {
		return "", fmt.Errorf("密文太短")
	}

	salt := data[:saltLen]
	key := pbkdf2.Key(masterKey, salt, iterCount, keyLen, sha256.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("创建cipher失败: %w", err)
	}

	ct := data[saltLen:]
	iv := ct[:aes.BlockSize]
	ct = ct[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ct, ct)

	return string(ct), nil
}
