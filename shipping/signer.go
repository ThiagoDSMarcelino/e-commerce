package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

type Signer struct {
	keysDir string
	priv    *rsa.PrivateKey
	pubs    map[string]*rsa.PublicKey
}

func NewSigner(settings *Settings) (*Signer, error) {
	priv, err := loadPrivateKey(settings.KeysDir, settings.ServiceName)
	if err != nil {
		return nil, err
	}

	return &Signer{
		keysDir: settings.KeysDir,
		priv:    priv,
		pubs:    map[string]*rsa.PublicKey{},
	}, nil
}

func (s *Signer) Sign(body []byte) (string, error) {
	hashed := sha256.Sum256(body)
	sig, err := rsa.SignPKCS1v15(rand.Reader, s.priv, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(sig), nil
}

func (s *Signer) Verify(from, sigB64 string, payload []byte) error {
	sig, err := base64.StdEncoding.DecodeString(sigB64)
	if err != nil {
		return fmt.Errorf("Malformed signature: %w", err)
	}

	pub, err := s.publicKey(from)
	if err != nil {
		return err
	}

	hashed := sha256.Sum256(payload)
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed[:], sig); err != nil {
		return fmt.Errorf("Invalid signature of %s: %w", from, err)
	}

	return nil
}

func (s *Signer) publicKey(service string) (*rsa.PublicKey, error) {
	pub, ok := s.pubs[service]
	if ok {
		return pub, nil
	}

	pub, err := loadPublicKey(s.keysDir, service)
	if err != nil {
		return nil, err
	}

	s.pubs[service] = pub

	return pub, nil
}
