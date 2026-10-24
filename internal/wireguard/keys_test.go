package wireguard

import (
	"encoding/base64"
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	if !ValidateKey(kp.PrivateKey) {
		t.Errorf("Invalid private key: %s", kp.PrivateKey)
	}

	if !ValidateKey(kp.PublicKey) {
		t.Errorf("Invalid public key: %s", kp.PublicKey)
	}

	// Verify public key derivation matches
	derivedPub, err := DerivePublicKey(kp.PrivateKey)
	if err != nil {
		t.Fatalf("DerivePublicKey failed: %v", err)
	}

	if derivedPub != kp.PublicKey {
		t.Errorf("Derived public key %s did not match generated public key %s", derivedPub, kp.PublicKey)
	}
}

func TestGeneratePresharedKey(t *testing.T) {
	psk, err := GeneratePresharedKey()
	if err != nil {
		t.Fatalf("GeneratePresharedKey failed: %v", err)
	}

	if !ValidateKey(psk) {
		t.Errorf("Invalid PSK: %s", psk)
	}
}

func TestValidateKey(t *testing.T) {
	// 32-byte valid key
	validKey := base64.StdEncoding.EncodeToString(make([]byte, 32))
	if !ValidateKey(validKey) {
		t.Errorf("Expected validKey to be valid")
	}

	// Invalid length
	shortKey := base64.StdEncoding.EncodeToString(make([]byte, 16))
	if ValidateKey(shortKey) {
		t.Errorf("Expected shortKey to be invalid")
	}

	// Invalid base64
	if ValidateKey("not-base64!!!") {
		t.Errorf("Expected invalid base64 to fail validation")
	}
}
