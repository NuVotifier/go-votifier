package votifier

import (
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