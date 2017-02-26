package votifier

import (
	"bytes"
	"crypto/rsa"
	"encoding/binary"
	"io"
	"log"
	"net"
	"time"
)

// Server represents a Votifier server.
type Server struct {
	listener    net.Listener
	voteHandler func(Vote)
	privateKey  *rsa.PrivateKey
	tokenFunc   ServiceTokenIdentifier
}

// NewServer creates a new Votifier server.
func NewServer(privateKey *rsa.PrivateKey, voteHandler func(Vote), tokenFunc ServiceTokenIdentifier) Server {
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

			challenge := randomString()
			c.SetDeadline(time.Now().Add(5 * time.Second))

			// Write greeting
			if _, err = io.WriteString(c, "VOTIFIER 2.0 "+challenge+"\n"); err != nil {
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

			if magicRead != v2Magic {
				v, err := deserializev1(data[:read], server.privateKey)
				if err != nil {
					log.Println(err)
					return
				}

				server.voteHandler(*v)
			} else {
				v, err := deserializev2(data[:read], server.tokenFunc, challenge)
				if err != nil {
					log.Println(err)
					return
				}

				server.voteHandler(*v)

				_, err = io.WriteString(c, "{\"status\":\"ok\"}")
				if err != nil {
					log.Println(err)
					return
				}
			}
		}(conn)
	}
}
