package crypto

import "testing"

func TestCryptoGCM(t *testing.T) {
	crp := GCM{
		Passphrase: "6368616e676520746869732070617373776f726420746f206120736563726574",
	}

	crp.Init()

	msg := "decentralized vpn"

	emsg, err := crp.Encrypt([]byte(msg))
	if err != nil {
		t.Error("unexpected error happened:", err)
	}

	dmsg, err := crp.Decrypt(emsg)
	if err != nil {
		t.Error("unexpected error happened:", err)
	}

	if string(dmsg) != msg {
		t.Errorf("expected %s but got, %s", msg, string(dmsg))
	}
}
