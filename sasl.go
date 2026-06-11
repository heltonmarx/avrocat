package main

import (
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"strings"

	"github.com/IBM/sarama"
	"github.com/xdg-go/scram"
)

type SASL struct {
	enabled   bool
	username  string
	password  string
	mechanism string
}

func (s *SASL) validate() error {
	switch {
	case !s.enabled:
		return nil
	case s.username == "":
		return errors.New("invalid kafka SASL username")
	case s.password == "":
		return errors.New("invalid kafka SASL password")
	case s.mechanism == "":
		return errors.New("invalid kafka SASL mechanism")
	}
	return nil
}

func ParseSASLMechanism(mech string) sarama.SASLMechanism {
	switch strings.ToUpper(mech) {
	case "PLAIN":
		return sarama.SASLTypePlaintext
	case "SCRAM-SHA-256":
		return sarama.SASLTypeSCRAMSHA256
	case "SCRAM-SHA-512":
		return sarama.SASLTypeSCRAMSHA512
	case "OAUTHBEARER":
		return sarama.SASLTypeOAuth
	case "GSSAPI":
		return sarama.SASLTypeGSSAPI
	default:
		// Defaults to PLAIN if unspecified or invalid
		return sarama.SASLTypePlaintext
	}
}

var (
	SHA256 scram.HashGeneratorFcn = sha256.New
	SHA512 scram.HashGeneratorFcn = sha512.New
)

type XDGSCRAMClient struct {
	*scram.Client
	*scram.ClientConversation
	scram.HashGeneratorFcn
}

func (x *XDGSCRAMClient) Begin(userName, password, authzID string) (err error) {
	x.Client, err = x.HashGeneratorFcn.NewClient(userName, password, authzID)
	if err != nil {
		return err
	}
	x.ClientConversation = x.Client.NewConversation()
	return nil
}

func (x *XDGSCRAMClient) Step(challenge string) (response string, err error) {
	response, err = x.ClientConversation.Step(challenge)
	return
}

func (x *XDGSCRAMClient) Done() bool {
	return x.ClientConversation.Done()
}
