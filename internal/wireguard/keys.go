package wireguard

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/curve25519"
)

// KeyPair represents a WireGuard Curve25519 public and private key pair.
type KeyPair struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

// GenerateKeyPair generates a new WireGuard-compatible Curve25519 key pair in Base64 format.
func GenerateKeyPair() (*KeyPair, error) {
	var privateKey [32]byte
	if _, err := io.ReadFull(rand.Reader, privateKey[:]); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Clamp the private key according to Curve25519 specification
	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64

	var publicKey [32]byte
	curve25519.ScalarBaseMult(&publicKey, &privateKey)

	return &KeyPair{
		PrivateKey: base64.StdEncoding.EncodeToString(privateKey[:]),
		PublicKey:  base64.StdEncoding.EncodeToString(publicKey[:]),
	}, nil
}

// DerivePublicKey derives the public key from a given Base64-encoded WireGuard private key.
func DerivePublicKey(privateKeyBase64 string) (string, error) {
	privBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return "", fmt.Errorf("invalid base64 private key: %w", err)
	}
	if len(privBytes) != 32 {
		return "", fmt.Errorf("private key must be exactly 32 bytes, got %d", len(privBytes))
	}

	var privateKey [32]byte
	copy(privateKey[:], privBytes)

	// Clamp to ensure consistency
	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64

	var publicKey [32]byte
	curve25519.ScalarBaseMult(&publicKey, &privateKey)

	return base64.StdEncoding.EncodeToString(publicKey[:]), nil
}

// GeneratePresharedKey generates a 256-bit symmetric pre-shared key (PSK) in Base64 format.
func GeneratePresharedKey() (string, error) {
	var psk [32]byte
	if _, err := io.ReadFull(rand.Reader, psk[:]); err != nil {
		return "", fmt.Errorf("failed to generate random psk bytes: %w", err)
	}
	return base64.StdEncoding.EncodeToString(psk[:]), nil
}

// ValidateKey checks if the base64 string decodes to 32 bytes.
func ValidateKey(keyBase64 string) bool {
	data, err := base64.StdEncoding.DecodeString(keyBase64)
	return err == nil && len(data) == 32
}
