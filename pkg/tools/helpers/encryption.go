package helpers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"sync"
)

var (
	instance *EncryptionHelper
	once     sync.Once
)

type EncryptionHelper struct {
	key []byte
	gcm cipher.AEAD
}

var (
	ErrEncryption  = errors.New("encryption failed")
	ErrDecryption  = errors.New("decryption failed")
	ErrInvalidKey  = errors.New("invalid encryption key")
	ErrInvalidData = errors.New("invalid data")
)

// GetInstance returns a singleton instance of EncryptionHelper
func GetEncryptionInstance() (*EncryptionHelper, error) {
	var err error
	once.Do(func() {
		instance, err = initialize()
	})
	if err != nil {
		return nil, err
	}
	return instance, nil
}

// initialize creates a new EncryptionHelper instance
func initialize() (*EncryptionHelper, error) {
	key := os.Getenv("ENCRYPTION_KEY_ENV")
	if key == "" {
		return nil, ErrInvalidKey
	}

	keyBytes := []byte(key)

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &EncryptionHelper{
		key: keyBytes,
		gcm: gcm,
	}, nil
}

// EncryptBytes encrypts byte array and returns base64 encoded result
func (eh *EncryptionHelper) EncryptBytes(data []byte) (string, error) {
	if len(data) == 0 {
		return "", ErrInvalidData
	}

	nonce := make([]byte, eh.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", ErrEncryption
	}

	sealed := eh.gcm.Seal(nonce, nonce, data, nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// DecryptBytes decrypts a base64 encoded encrypted data to bytes
func (eh *EncryptionHelper) DecryptBytes(encryptedData string) ([]byte, error) {
	if encryptedData == "" {
		return nil, ErrInvalidData
	}

	decoded, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, ErrDecryption
	}

	if len(decoded) < eh.gcm.NonceSize() {
		return nil, ErrDecryption
	}

	nonce := decoded[:eh.gcm.NonceSize()]
	ciphertext := decoded[eh.gcm.NonceSize():]

	plaintext, err := eh.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryption
	}

	return plaintext, nil
}
