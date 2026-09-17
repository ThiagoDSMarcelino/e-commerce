package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

type Verifier struct {
	keysDir string
	pubs    map[string]*rsa.PublicKey
}

func NewVerifier(settings *Settings) *Verifier {
	return &Verifier{
		keysDir: settings.KeysDir,
		pubs:    map[string]*rsa.PublicKey{},
	}
}

func (v *Verifier) Verify(from, sigB64 string, payload []byte) error {
	sig, err := base64.StdEncoding.DecodeString(sigB64)
	if err != nil {
		return fmt.Errorf("Malformed signature: %w", err)
	}

	pub, err := v.publicKey(from)
	if err != nil {
		return err
	}

	hashed := sha256.Sum256(payload)
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed[:], sig); err != nil {
		return fmt.Errorf("Invalid signature of %s: %w", from, err)
	}

	return nil
}

func (v *Verifier) publicKey(service string) (*rsa.PublicKey, error) {
	pub, ok := v.pubs[service]
	if ok {
		return pub, nil
	}

	pub, err := loadPublicKey(v.keysDir, service)
	if err != nil {
		return nil, err
	}

	v.pubs[service] = pub

	return pub, nil
}
