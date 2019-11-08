package crypto

import (
	"crypto/aes"
	"testing"
)

func TestCryptoCBC(t *testing.T) {
	c := &CBC{
		Passphrase: "6368616e676520746869732070617373776f726420746f206120736563726574",
	}

	c.Init()

	msg := "decentralized vpn"

	emsg, err := c.Encrypt([]byte(msg))
	if err != nil {
		t.Error("unexpected error happened:", err)
	}

	dmsg, err := c.Decrypt(emsg)
	if err != nil {
		t.Error("unexpected error happened:", err)
	}

	if string(dmsg) != msg {
		t.Errorf("expected %s but got, %s", msg, string(dmsg))
	}
}

func TestPadding(t *testing.T) {
	msg := "vpn"
	b := padding([]byte(msg))

	if len(b) < 1 {
		t.Error("unexpected size length")
	}

	padLen := int(b[len(b)-1])
	if (aes.BlockSize - 3) != padLen {
		t.Errorf("expected padding length %d but got, %d", aes.BlockSize-3, padLen)
	}

	for _, e := range b[len(b)-padLen:] {
		if int(e) != padLen {
			t.Error("unexpected repeated padding element")
		}
	}
}

func TestUnpadding(t *testing.T) {
	msg := "vpn"
	b := padding([]byte(msg))
	ub, err := unpadding(b)
	if err != nil {
		t.Error("unexpected error", err)
	}

	if string(ub) != "vpn" || len(ub) > 3 {
		t.Error("unexpected padded result")
	}
}
