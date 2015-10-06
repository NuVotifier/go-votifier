package votifier

import (
	"crypto/rsa"
	"io"
	"log"
	"net"
)

type Server struct {
	listener    net.Listener
	voteHandler func(Vote)
	privateKey  *rsa.PrivateKey
}

func NewServer(privateKey *rsa.PrivateKey, voteHandler func(Vote)) Server {
	return Server{privateKey: privateKey, voteHandler: voteHandler}
}

func (server Server) Close() {
	server.listener.Close()
}

func (server Server) ListenAndServe(address string) {
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

			// Write greeting
			_, err := io.WriteString(c, "VOTIFIER 1.9\r\n")
			if err != nil {
				log.Println(err)
				return
			}

			// Read in what data we can and try to handle it
			data := make([]byte, 256)
			_, err = c.Read(data)
			if err != nil {
				log.Println(err)
				return
			}

			v, err := Deserialize(data, server.privateKey)
			if err != nil {
				log.Println(err)
				return
			}

			server.voteHandler(*v)
		}(conn)
	}
}
