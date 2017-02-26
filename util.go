package votifier

import (
	"math/rand"
	"time"
)

const alpha = "abcdefghijklmnopqrstuvwxyz0123456789"

func randomString() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	p := make([]byte, 24)
	for i := range p {
		p[i] = alpha[r.Intn(len(alpha))]
	}
	return string(p)
}

// StaticServiceTokenIdentifier returns a ServiceTokenIdentifier that returns a static token for any service.
func StaticServiceTokenIdentifier(token string) ServiceTokenIdentifier {
	return func(string) string {
		return token
	}
}
