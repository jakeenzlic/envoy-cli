package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"os"
)

// Vault holds encrypted environment variables.
type Vault struct {
	Entries map[string]string `json:"entries"`
}

// New creates an empty Vault.
func New() *Vault {
	return &Vault{Entries: make(map[string]string)}
}

// Set adds or updates a key-value pair in the vault.
func (v *Vault) Set(key, value string) {
	v.Entries[key] = value
}

// Get retrieves a value by key.
func (v *Vault) Get(key string) (string, bool) {
	val, ok := v.Entries[key]
	return val, ok
}

// Delete removes a key from the vault.
func (v *Vault) Delete(key string) {
	delete(v.Entries, key)
}

// deriveKey produces a 32-byte AES key from a passphrase.
func deriveKey(passphrase string) []byte {
	hash := sha256.Sum256([]byte(passphrase))
	return hash[:]
}

// Save encrypts and writes the vault to a file.
func (v *Vault) Save(path, passphrase string) error {
	plaintext, err := json.Marshal(v)
	if err != nil {
		return err
	}
	block, err := aes.NewCipher(deriveKey(passphrase))
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return os.WriteFile(path, ciphertext, 0600)
}

// Load decrypts and reads the vault from a file.
func Load(path, passphrase string) (*Vault, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(deriveKey(passphrase))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("vault: ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("vault: decryption failed — wrong passphrase?")
	}
	v := &Vault{}
	if err := json.Unmarshal(plaintext, v); err != nil {
		return nil, err
	}
	return v, nil
}
