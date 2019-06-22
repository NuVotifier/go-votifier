package votifier

import (
	"bytes"
	"crypto/rsa"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"net"
	"runtime"
	"time"
)

type VotifierProtocol int

const (
	VotifierV1 VotifierProtocol = iota
	VotifierV2 VotifierProtocol = iota
)

// VoteListener takes a vote and an int describing the protocol version (1 or 2).
type VoteListener func(Vote, VotifierProtocol, interface{})

type ReceiverRecord struct {
	PrivateKey *rsa.PrivateKey        // v1
	TokenId    ServiceTokenIdentifier // v2
	Metadata   interface{}
}

// Server represents a Votifier server.
type Server struct {
	listener    net.Listener
	voteHandler VoteListener
	records     []ReceiverRecord
}

// NewServer creates a new Votifier server.
func NewServer(voteHandler VoteListener, records []ReceiverRecord) Server {
	return Server{
		voteHandler: voteHandler,
		records:     records,
	}
}

// ListenAndServe binds to a specified address-port pair and starts serving Votifier requests.
func (server *Server) ListenAndServe(address string) error {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	return server.Serve(l)
}

// Serve serves requests on the provided listener.
func (server *Server) Serve(l net.Listener) error {
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		// Handle the connection.
		go func(c net.Conn) {
			defer c.Close()

			challenge, err := randomString()
			if err != nil {
				// something very bad happened - only caused when /dev/urandom
				// also returns an error, which should never happen.
				log.Println(err)
				return
			}
			c.SetDeadline(time.Now().Add(5 * time.Second))

			// Write greeting
			if _, err = io.WriteString(c, "VOTIFIER 2 "+challenge+"\n"); err != nil {
				log.Println(err)
				return
			}

			// Read in what data we can and try to handle it
			data := make([]byte, 1024)
			read, err := c.Read(data)
			if err != nil {
				log.Println(err)
				return
			}

			// Do we have v2 magic?
			reader := bytes.NewReader(data[:2])
			var magicRead int16
			if err = binary.Read(reader, binary.BigEndian, &magicRead); err != nil {
				log.Println(err)
				return
			}

			isv2 := magicRead == v2Magic
			defer func() {
				if rerr := recover(); rerr != nil {
					if _, ok := rerr.(runtime.Error); ok {
						panic(rerr)
					}
					frerr := rerr.(error)
					if isv2 {
						result := v2Response{
							Status: "error",
							Error:  frerr.Error(),
							Cause:  "panic",
						}
						json.NewEncoder(c).Encode(result)
					}
				}
			}()

			for _, record := range server.records {
				if !isv2 && record.PrivateKey != nil {
					v, err := deserializev1(data[:read], record.PrivateKey)
					if err != nil {
						continue
					}

					server.voteHandler(*v, VotifierV1, record.Metadata)
					return
				} else {
					v, err := deserializev2(data[:read], record.TokenId, challenge)
					if err != nil {
						continue
					}

					server.voteHandler(*v, VotifierV2, record.Metadata)

					io.WriteString(c, "{\"status\":\"ok\"}")
					return
				}
			}

			// We couldn't decrypt it correctly
			if isv2 {
				result := v2Response{
					Status: "error",
					Error:  err.Error(),
					Cause:  "decode",
				}
				json.NewEncoder(c).Encode(result)
			}
		}(conn)
	}
}
