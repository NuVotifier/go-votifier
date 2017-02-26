package votifier

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"time"
)

// V2Client represents a Votifier v2 client.
type V2Client struct {
	address string
	token   string
}

type v2Response struct {
	Status string
	Cause  string
	Error  string
}

// NewV2Client creates a new Votifier v2 client.
func NewV2Client(address string, token string) *V2Client {
	return &V2Client{address, token}
}

// SendVote sends a vote through the client.
func (client *V2Client) SendVote(vote Vote) error {
	conn, err := net.DialTimeout("tcp", client.address, 3*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(3 * time.Second))

	greeting := make([]byte, 64)
	read, err := conn.Read(greeting)
	if err != nil {
		return err
	}

	parts := bytes.Split(greeting[:read-1], []byte(" "))
	if len(parts) != 3 {
		return errors.New("not a v2 server")
	}
	challenge := string(parts[2])

	serialized, err := vote.serializev2(client.token, challenge)
	if err != nil {
		return err
	}
	_, err = conn.Write(*serialized)

	// read response
	responseBuf := make([]byte, 256)
	read, err = conn.Read(responseBuf)
	if err != nil {
		return err
	}

	var response v2Response
	if err := json.NewDecoder(bytes.NewBuffer(responseBuf[:read])).Decode(&response); err != nil {
		return nil
	}

	if response.Status == "ok" {
		return nil
	}

	return errors.New("remote server error: " + response.Cause + ": " + response.Error)
}
