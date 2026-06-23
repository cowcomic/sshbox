package crypto

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	masterKey := []byte("0123456789abcdef0123456789abcdef") // 32 bytes
	plaintext := "my-secret-password"

	encrypted, err := Encrypt(plaintext, masterKey)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	if encrypted == plaintext {
		t.Error("encrypted text should differ from plaintext")
	}

	decrypted, err := Decrypt(encrypted, masterKey)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Decrypt() = %q, want %q", decrypted, plaintext)
	}
}

func TestEncryptDifferentResults(t *testing.T) {
	masterKey := []byte("0123456789abcdef0123456789abcdef")
	plaintext := "same-password"

	e1, _ := Encrypt(plaintext, masterKey)
	e2, _ := Encrypt(plaintext, masterKey)

	if e1 == e2 {
		t.Error("two encryptions of same plaintext should differ (random salt/iv)")
	}
}

func TestDecryptWrongKey(t *testing.T) {
	key1 := []byte("0123456789abcdef0123456789abcdef")
	key2 := []byte("fedcba9876543210fedcba9876543210")

	encrypted, _ := Encrypt("secret", key1)
	decrypted, err := Decrypt(encrypted, key2)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}
	// CFB mode doesn't fail on wrong key, but produces garbage
	if decrypted == "secret" {
		t.Error("Decrypt with wrong key should not return original plaintext")
	}
}

func TestDecryptInvalidBase64(t *testing.T) {
	_, err := Decrypt("not-valid-base64!!!", []byte("key"))
	if err == nil {
		t.Error("Decrypt with invalid base64 should fail")
	}
}

func TestDecryptTooShort(t *testing.T) {
	// Valid base64 but too short
	_, err := Decrypt("YWJj", []byte("0123456789abcdef0123456789abcdef"))
	if err == nil {
		t.Error("Decrypt with too-short ciphertext should fail")
	}
}
