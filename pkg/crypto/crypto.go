package crypto

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/pkg/errors"
)

// KeyPair is a simple struct used for returning a key pair from the function
// defined in this module that creates a key pair. We pass back the generated
// keypair to be stored and passed down the stack in order to encrypt data for
// the intended recipient.
type KeyPair struct {
	// PrivateKey is the private key part of the key pair.
	PrivateKey string

	// PublicKey is the public key part of the key pair.
	PublicKey string
}

// NewKeyPair is a function that returns an instantiated key pair ready for
// persistence to the DB. Currently this is a fake implementation that just
// reads some random bytes /dev/urandom, so we should re-implement this function
// using Zenroom, to create a real key pair that Zenroom is able to use for
// encryption.
func NewKeyPair() (*KeyPair, error) {
	priv := make([]byte, 32)
	_, err := rand.Read(priv)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create fake private key")
	}

	pub := make([]byte, 32)
	_, err = rand.Read(pub)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create fake public key")
	}

	return &KeyPair{
		PrivateKey: base64.StdEncoding.EncodeToString(priv),
		PublicKey:  base64.StdEncoding.EncodeToString(pub),
	}, nil
}
