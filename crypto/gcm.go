package crypto

import(
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"
	"log"
)

// GCM represents Galois/Counter Mode 
type GCM struct {
	Passphrase string
}

// Encrypt encrypts the plaindata
func (g GCM) Encrypt(plainData []byte) []byte {
	key, _ := hex.DecodeString(g.Passphrase)
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatal(err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.Fatal(err)
	}

	return gcm.Seal(nonce, nonce, plainData, nil)
}

// Decrypt decrypts the cipherdata
func (g GCM) Decrypt(cipherData []byte) []byte {
	key, _ := hex.DecodeString(g.Passphrase)
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatal(err)
	}

	nonceSize := gcm.NonceSize()
	nonce, cipherData := cipherData[:nonceSize], cipherData[nonceSize:]
	plainData, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		log.Fatal(err)
	}

	return plainData
}