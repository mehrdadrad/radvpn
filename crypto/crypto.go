package crypto

import (
	"crypto/sha1"
	"errors"
)

// Cipher interfaces to different cryptographies
type Cipher interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)
	Init()
}

// Pbkdf1 applies a hash function, which shall be SHA-1 to derive keys
// tools.ietf.org/html/rfc8018#section-5
func Pbkdf1(pass, salt string, count, dkLen int) ([]byte, error) {
	if dkLen > 20 {
		return nil, errors.New("derived key too long")
	}

	derived := make([]byte, len(pass)+len(salt))
	copy(derived, pass)
	copy(derived[len(pass):], salt)

	for i := 0; i < count; i++ {
		d := sha1.Sum(derived)
		derived = d[:]
	}

	return derived[:dkLen], nil
}
