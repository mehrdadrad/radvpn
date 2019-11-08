package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"
)

// GCM represents Galois/Counter Mode
type GCM struct {
	Passphrase string
	key        []byte
}

// Init initializes the key based on the passphrase
func (g *GCM) Init() {
	g.key, _ = hex.DecodeString(g.Passphrase)
}

// Encrypt encrypts the plaindata
func (g GCM) Encrypt(plainData []byte) ([]byte, error) {
	block, err := aes.NewCipher(g.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plainData, nil), nil
}

// Decrypt decrypts the cipherdata
func (g GCM) Decrypt(cipherData []byte) ([]byte, error) {
	block, err := aes.NewCipher(g.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	nonce, cipherData := cipherData[:nonceSize], cipherData[nonceSize:]
	plainData, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return nil, err
	}

	return plainData, nil
}
