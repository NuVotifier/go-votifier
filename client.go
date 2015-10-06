package votifier

import (
	"crypto/rsa"
	"net"
)

type Client struct {
	Address   string
	PublicKey *rsa.PublicKey
}

func NewClient(address string, publicKey *rsa.PublicKey) Client {
	return Client{address, publicKey}
}

// Sends a vote.
func (client Client) SendVote(vote Vote) error {
	conn, err := net.Dial("tcp", client.Address)
	if err != nil {
		return err
	}
	defer conn.Close()

	serialized, err := vote.Serialize(client.PublicKey)
	if err != nil {
		return err
	}

	_, err = conn.Write(*serialized)
	return err
}
