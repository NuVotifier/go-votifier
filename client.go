package votifier

import (
	"crypto/rsa"
	"net"
	"time"
)

// Client represents a Votifier client.
type Client struct {
	address   string
	publicKey *rsa.PublicKey
}

// NewClient creates a new Votifier client.
func NewClient(address string, publicKey *rsa.PublicKey) Client {
	return Client{address, publicKey}
}

// SendVote sends a vote through the client.
func (client Client) SendVote(vote Vote) error {
	conn, err := net.DialTimeout("tcp", client.address, 3*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(3 * time.Second))
	serialized, err := vote.Serialize(client.publicKey)
	if err != nil {
		return err
	}

	_, err = conn.Write(*serialized)
	return err
}
