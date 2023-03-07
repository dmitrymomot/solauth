package solauth

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"

	"github.com/mr-tron/base58"
	"github.com/pkg/errors"
)

// VerifySignature verifies the signature of the request.
// This function verifies the signature of the message using
// the public key of the sender.
// It returns error if the signature is NOT valid, otherwise nil.
func VerifySignature(message, signature, publicKey string) error {
	publicKeyBytes, err := base58.Decode(publicKey)
	if err != nil {
		return fmt.Errorf("failed to decode public key: %w", err)
	}

	if len(publicKeyBytes) != ed25519.PublicKeySize {
		return errors.Errorf("expected ed25519 public key size is: %v, got: %v", ed25519.PublicKeySize, len(publicKeyBytes))
	}

	ed25519PublicKey := ed25519.PublicKey(publicKeyBytes)
	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return errors.Wrap(err, "can't decode base64 signature")
	}

	validSignature := ed25519.Verify(ed25519PublicKey, []byte(message), sig)
	if !validSignature {
		return errors.Errorf("signature is incorrect")
	}

	return nil
}
