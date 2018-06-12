package crypto

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/thingful/zenroom-go"

	"github.com/thingful/iotdevicereg/pkg/lua"
)

// KeyPair is a simple struct used for returning a key pair from the function
// defined in this module that creates a key pair. We pass back the generated
// keypair to be stored and passed down the stack in order to encrypt data for
// the intended recipient.
type KeyPair struct {
	// PrivateKey is the private key part of the key pair.
	PrivateKey string `json:"secret"`

	// PublicKey is the public key part of the key pair.
	PublicKey string `json:"public"`
}

// NewKeyPair is a function that returns an instantiated key pair ready for
// persistence to the DB. Currently this is a fake implementation that just
// reads some random bytes /dev/urandom, so we should re-implement this function
// using Zenroom, to create a real key pair that Zenroom is able to use for
// encryption.
func NewKeyPair() (*KeyPair, error) {
	script, err := lua.Asset("generatekeys.lua")
	if err != nil {
		return nil, errors.Wrap(err, "failed to read zenroom script")
	}

	keys, err := zenroom.Exec(string(script), "", "")
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute zenroom script")
	}

	var keyPair KeyPair
	err = json.Unmarshal([]byte(keys), &keyPair)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal zenroom output")
	}

	return &keyPair, nil
}
