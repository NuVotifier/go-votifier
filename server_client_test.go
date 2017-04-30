package votifier

import (
	"crypto/rsa"
	"errors"
	"net"
	"strings"
	"testing"
)

func TestServer(t *testing.T) {
	v := NewVote("golang", "golang", "127.0.0.1")

	// Generate a set of keys for later use
	key, err := rsa.GenerateKey(new(badRandomReader), 2048)
	if err != nil {
		t.Error(err)
		return
	}

	for i := 1; i <= 2; i++ {
		vl := func(rv Vote, ver int) {
			if rv != v {
				t.Error("Vote received did not match original")
			}

			if ver != i {
				t.Error("Vote is not v" + string(i))
			}
		}

		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Error(err)
		}
		defer listener.Close()

		var client Client
		if i == 1 {
			pk := key.PublicKey
			client = NewV1Client(listener.Addr().String(), &pk)
		} else {
			client = NewV2Client(listener.Addr().String(), "abcxyz")
		}
		server := NewServer(key, vl, StaticServiceTokenIdentifier("abcxyz"))
		go server.Serve(listener)

		err = client.SendVote(v)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestServerv2Panic(t *testing.T) {
	vl := func(rv Vote, ver int) {
		panic(errors.New("boom"))
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Error(err)
	}
	defer listener.Close()
	server := NewServer(nil, vl, StaticServiceTokenIdentifier("abcxyz"))
	go server.Serve(listener)

	client := NewV2Client(listener.Addr().String(), "abcxyz")
	err = client.SendVote(NewVote("golang", "golang", "127.0.0.1"))
	if err == nil {
		t.Error("expected error, but didn't get any")
	}

	if !strings.HasSuffix(err.Error(), "panic: boom") {
		t.Error("invalid error from error")
	}
}
