package votifier

// Client represents a Votifier client.
type Client interface {
	// SendVote sends a vote through the client.
	SendVote(vote Vote) error
}
