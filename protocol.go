package votifier

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"strings"
	"time"
)

type Vote struct {
	// The name of the service the user is voting from.
	ServiceName string

	// The user's Minecraft username.
	Username string

	// The voting user's IP address.
	Address string

	// The timestamp this vote was issued.
	Timestamp string
}

// Creates a new vote, filling in the timestamp.
func NewVote(serviceName string, username string, address string) Vote {
	return Vote{serviceName, username, address, time.Now().String()}
}

func Deserialize(msg []byte, privateKey *rsa.PrivateKey) (*Vote, error) {
	decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, msg)
	if err != nil {
		return nil, err
	}

	elements := strings.Split(string(decrypted), "\n")
	if len(elements) != 6 {
		return nil, fmt.Errorf("Element count is invalid; wanted 6, got %d", len(elements))
	}
	if elements[0] != "VOTE" {
		return nil, fmt.Errorf("First element is incorrect; expected 'VOTE', got %s", elements[0])
	}
	return &Vote{elements[1], elements[2], elements[3], elements[4]}, nil
}

// Serializes the vote.
func (vote Vote) Serialize(publicKey *rsa.PublicKey) (*[]byte, error) {
	s := strings.Join([]string{"VOTE", vote.ServiceName, vote.Username, vote.Address, vote.Timestamp, ""}, "\n")
	msg := []byte(s)

	// Encrypt the vote using the supplied public key.
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, msg)
	if err != nil {
		return nil, err
	}

	return &encrypted, nil
}
