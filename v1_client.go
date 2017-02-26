package votifier

import (
	"crypto/rsa"
	"net"
	"time"
)

// V1Client represents a Votifier v1 client.
type V1Client struct {
	address   string
	publicKey *rsa.PublicKey
}

// NewV1Client creates a new Votifier client.
func NewV1Client(address string, publicKey *rsa.PublicKey) *V1Client {
	return &V1Client{address, publicKey}
}

// SendVote sends a vote through the client.
func (client *V1Client) SendVote(vote Vote) error {
	conn, err := net.DialTimeout("tcp", client.address, 3*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(3 * time.Second))
	serialized, err := vote.serializev1(client.publicKey)
	if err != nil {
		return err
	}

	_, err = conn.Write(*serialized)
	return err
}
