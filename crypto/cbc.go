package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
)

// CBC represents block cipher mode of operation algorithm
type CBC struct {
	Passphrase string
	key        []byte
}

// Init initializes the key based on the passphrase
func (c *CBC) Init() {
	c.key, _ = hex.DecodeString(c.Passphrase)
}

// Encrypt encrypts the plaindat
func (c CBC) Encrypt(plainData []byte) ([]byte, error) {
	if len(plainData)%aes.BlockSize != 0 {
		plainData = padding(plainData)
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	cipherData := make([]byte, aes.BlockSize+len(plainData))
	iv := cipherData[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherData[aes.BlockSize:], plainData)

	return cipherData, nil
}

// Decrypt decrypts the cipherdat
func (c CBC) Decrypt(cipherData []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	if len(cipherData) < aes.BlockSize {
		return nil, errors.New("encrypted data is too short")
	}

	iv := cipherData[:aes.BlockSize]
	cipherData = cipherData[aes.BlockSize:]

	if len(cipherData)%aes.BlockSize != 0 {
		return nil, errors.New("invalid size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(cipherData, cipherData)

	cipherData, _ = unpadding(cipherData)

	return cipherData, nil
}

func padding(b []byte) []byte {
	padLen := aes.BlockSize - (len(b) % aes.BlockSize)
	pad := bytes.Repeat([]byte{byte(padLen)}, padLen)

	return append(b, pad...)
}

func unpadding(b []byte) ([]byte, error) {
	bLen := len(b)

	if bLen < 1 {
		return b, errors.New("invalid size")
	}

	pad := b[len(b)-1]
	padLen := int(pad)

	if padLen > bLen || padLen > aes.BlockSize {
		return b, errors.New("invalid padding size")
	}

	for _, p := range b[bLen-padLen : bLen-1] {
		if p != pad {
			return b, errors.New("invalid padding")
		}
	}

	return b[:bLen-padLen], nil
}
