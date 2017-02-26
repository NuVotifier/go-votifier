package votifier

import (
	"crypto/rsa"
	"io"
	"log"
	"net"
	"time"
)

type Server struct {
	listener    net.Listener
	voteHandler func(Vote)
	privateKey  *rsa.PrivateKey
	tokenFunc   ServiceTokenIdentifier
}

func NewServer(privateKey *rsa.PrivateKey, voteHandler func(Vote), tokenFunc ServiceTokenIdentifier) Server {
	return Server{privateKey: privateKey, voteHandler: voteHandler,
		tokenFunc: tokenFunc}
}

func (server *Server) Close() {
	server.listener.Close()
}

func (server *Server) ListenAndServe(address string) {
	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Println(err)
		return
	}
	server.listener = l
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			break
		}

		// Handle the connection.
		go func(c net.Conn) {
			defer c.Close()

			challenge := randomString()
			c.SetDeadline(time.Now().Add(5 * time.Second))

			// Write greeting
			_, err = io.WriteString(c, "VOTIFIER 2.0 "+challenge+"\r\n")
			if err != nil {
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

			if read == 256 {
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
