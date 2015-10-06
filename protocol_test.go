package votifier

import (
	"crypto/rsa"
	"math/rand"
	"testing"
)

// Extremely bad random number generation, used only for testing purposes.
type badRandomReader struct{}

func (badRandomReader) Read(p []byte) (n int, err error) {
	for i, _ := range p {
		p[i] = byte(rand.Intn(255))
	}
	return len(p), nil
}

func isEq(expected string, got string, n string, t *testing.T) {
	if expected != got {
		t.Error("Field %s doesn't match; expected '%s', got '%s'", n, expected, got)
	}
}

func TestSerialization(t *testing.T) {
	v := NewVote("golang", "golang", "127.0.0.1")

	// Generate a set of keys for later use
	key, err := rsa.GenerateKey(new(badRandomReader), 2048)
	if err != nil {
		t.Error(err)
	}

	// Try to encrypt this vote.
	s, err := v.Serialize(&key.PublicKey)
	if err != nil {
		t.Error(err)
	}

	if len(*s) != 256 {
		t.Error("Encrypted PKCS1v15 output should be 256 bytes, but it is %d bytes long", len(*s))
	}

	// Try to decrypt this vote.
	d, err := Deserialize(*s, key)
	if err != nil {
		t.Error(err)
	}

	isEq(v.ServiceName, d.ServiceName, "ServiceName", t)
	isEq(v.Username, d.Username, "Username", t)
	isEq(v.Address, d.Address, "Address", t)
	isEq(v.Timestamp, d.Timestamp, "Timestamp", t)
}
