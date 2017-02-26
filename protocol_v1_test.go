package votifier

import (
	"crypto/rsa"
	"math/rand"
	"testing"
)

// Extremely bad random number generation, used only for testing purposes.
type badRandomReader struct{}

func (badRandomReader) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = byte(rand.Intn(255))
	}
	return len(p), nil
}

func TestSerializationv1(t *testing.T) {
	v := NewVote("golang", "golang", "127.0.0.1")

	// Generate a set of keys for later use
	key, err := rsa.GenerateKey(new(badRandomReader), 2048)
	if err != nil {
		t.Error(err)
		return
	}

	// Try to encrypt this vote.
	s, err := v.serializev1(&key.PublicKey)
	if err != nil {
		t.Error(err)
		return
	}

	if len(*s) != 256 {
		t.Error("Encrypted PKCS1v15 output should be 256 bytes, but it is %d bytes long", len(*s))
		return
	}

	// Try to decrypt this vote.
	d, err := deserializev1(*s, key)
	if err != nil {
		t.Error(err)
		return
	}

	if v != *d {
		t.Error("Votes don't match")
		return
	}
}
