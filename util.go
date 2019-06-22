package votifier

import (
	"crypto/rand"
	"encoding/base64"
)

func randomString() (string, error) {
	p := make([]byte, 24)
	_, err := rand.Read(p)
	if err != nil {
		return "", err
	}

	return base64.RawStdEncoding.EncodeToString(p), nil
}

// StaticServiceTokenIdentifier returns a ServiceTokenIdentifier that returns a static token for any service.
func StaticServiceTokenIdentifier(token string) ServiceTokenIdentifier {
	return func(string) string {
		return token
	}
}
