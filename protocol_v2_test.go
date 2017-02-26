package votifier

import "testing"

func TestSerializationv2(t *testing.T) {
	v := NewVote("golang", "golang", "127.0.0.1")

	// Try to encrypt this vote.
	s, err := v.serializev2("abcxyz", "xyz")
	if err != nil {
		t.Error(err)
		return
	}

	// Try to decrypt this vote.
	d, err := deserializev2(*s, StaticServiceTokenIdentifier("abcxyz"), "xyz")
	if err != nil {
		t.Error(err)
		return
	}

	if v != *d {
		t.Error("Votes don't match: ", v, "-", *d)
	}
}
