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
type VoteListener func(Vote, VotifierProtocol)

// Server represents a Votifier server.
type Server struct {
	listener    net.Listener
	voteHandler VoteListener
	privateKey  *rsa.PrivateKey
	tokenFunc   ServiceTokenIdentifier
}

// NewServer creates a new Votifier server.
func NewServer(privateKey *rsa.PrivateKey, voteHandler VoteListener, tokenFunc ServiceTokenIdentifier) Server {
	return Server{privateKey: privateKey, voteHandler: voteHandler,
		tokenFunc: tokenFunc}
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

			if !isv2 && server.privateKey != nil {
				v, err := deserializev1(data[:read], server.privateKey)
				if err != nil {
					log.Println(err)
					return
				}

				server.voteHandler(*v, VotifierV1)
			} else {
				v, err := deserializev2(data[:read], server.tokenFunc, challenge)
				if err != nil {
					log.Println(err)
					result := v2Response{
						Status: "error",
						Error:  err.Error(),
						Cause:  "decode",
					}
					json.NewEncoder(c).Encode(result)
					return
				}

				server.voteHandler(*v, VotifierV2)

				_, err = io.WriteString(c, "{\"status\":\"ok\"}")
				if err != nil {
					log.Println(err)
					return
				}
			}
		}(conn)
	}
}
