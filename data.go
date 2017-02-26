package votifier

import (
	"strconv"
	"time"
)

// Vote represents a Votifier vote.
type Vote struct {
	// The name of the service the user is voting from.
	ServiceName string `json:"serviceName"`

	// The user's Minecraft username.
	Username string `json:"username"`

	// The voting user's IP address.
	Address string `json:"address"`

	// The timestamp this vote was issued.
	Timestamp string `json:"timeStamp"`
}

// NewVote creates a new vote and pre-fills the timestamp.
func NewVote(serviceName string, username string, address string) Vote {
	return Vote{serviceName, username, address, strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)}
}
