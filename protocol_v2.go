package votifier

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

const (
	v2Magic int16 = 0x733A
)

type votifier2Wrapper struct {
	Payload   string `json:"payload"`
	Signature []byte `json:"signature"`
}

type votifier2Inner struct {
	ServiceName string `json:"serviceName"`
	Username    string `json:"username"`
	Address     string `json:"address"`
	Timestamp   int64  `json:"timestamp"`
	Challenge   string `json:"challenge"`
}

// ServiceTokenIdentifier defines a function for identifying a token for a service.
type ServiceTokenIdentifier func(string) string

func deserializev2(msg []byte, tokenFunc ServiceTokenIdentifier, challenge string) (*Vote, error) {
	reader := bytes.NewReader(msg)

	// verify v2 magic
	var magicRead int16
	err := binary.Read(reader, binary.BigEndian, &magicRead)
	if err != nil {
		return nil, err
	}

	if magicRead != v2Magic {
		return nil, errors.New("v2 magic mismatch")
	}

	// read message length
	var bytes int16
	if err = binary.Read(reader, binary.BigEndian, &bytes); err != nil {
		return nil, err
	}

	// now for the fun part
	var wrapper votifier2Wrapper
	if err = json.NewDecoder(reader).Decode(&wrapper); err != nil {
		return nil, err
	}

	var vote votifier2Inner
	if err = json.NewDecoder(strings.NewReader(wrapper.Payload)).Decode(&vote); err != nil {
		return nil, err
	}

	// validate challenge
	if vote.Challenge != challenge {
		return nil, errors.New("challenge invalid")
	}

	// validate HMAC
	m := hmac.New(sha256.New, []byte(tokenFunc(vote.ServiceName)))
	m.Write([]byte(wrapper.Payload))
	s := m.Sum(nil)
	if !hmac.Equal(s, wrapper.Signature) {
		return nil, errors.New("signature invalid")
	}

	return &Vote{
		ServiceName: vote.ServiceName,
		Address:     vote.Address,
		Username:    vote.Username,
		Timestamp:   strconv.FormatInt(vote.Timestamp, 10),
	}, nil
}

func (v Vote) serializev2(token string, challenge string) (*[]byte, error) {
	ts, err := strconv.ParseInt(v.Timestamp, 10, 64)
	if err != nil {
		// do our best
		ts = 0
	}
	inner := votifier2Inner{
		ServiceName: v.ServiceName,
		Address:     v.Address,
		Username:    v.Username,
		Timestamp:   ts,
		Challenge:   challenge,
	}

	// encode inner vote and generate outer package
	var innerBuf bytes.Buffer
	if err = json.NewEncoder(&innerBuf).Encode(inner); err != nil {
		return nil, err
	}

	innerJSON := innerBuf.String()
	m := hmac.New(sha256.New, []byte(token))
	innerBuf.WriteTo(m)

	wrapper := votifier2Wrapper{
		Payload:   innerJSON,
		Signature: m.Sum(nil),
	}

	// assemble full package
	var wrapperBuf bytes.Buffer
	if err = json.NewEncoder(&wrapperBuf).Encode(wrapper); err != nil {
		return nil, err
	}

	var finalBuf bytes.Buffer
	if err = binary.Write(&finalBuf, binary.BigEndian, v2Magic); err != nil {
		return nil, err
	}
	if err = binary.Write(&finalBuf, binary.BigEndian, int16(wrapperBuf.Len())); err != nil {
		return nil, err
	}
	wrapperBuf.WriteTo(&finalBuf)
	finalBytes := finalBuf.Bytes()
	return &finalBytes, nil
}
