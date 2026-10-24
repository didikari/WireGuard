package client

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSanitizeProfileName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"my-vpn.conf", "my-vpn"},
		{"office_vpn", "office_vpn"},
		{"home/vpn!@#$", "homevpn"},
		{"  client-1  ", "client-1"},
	}

	for _, tt := range tests {
		got := sanitizeProfileName(tt.input)
		if got != tt.expected {
			t.Errorf("sanitizeProfileName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestImportAndListProfiles(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ezwg-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cm := &ClientManager{
		ProfilesDir: tempDir,
	}

	sampleConf := `[Interface]
PrivateKey = YWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWE=
Address = 10.8.0.2/32
DNS = 1.1.1.1

[Peer]
PublicKey = YmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmI=
Endpoint = 203.0.113.1:51820
AllowedIPs = 0.0.0.0/0
`

	// Import direct content
	savedPath, err := cm.ImportProfile("test-vpn", sampleConf)
	if err != nil {
		t.Fatalf("ImportProfile failed: %v", err)
	}

	expectedPath := filepath.Join(tempDir, "test-vpn.conf")
	if savedPath != expectedPath {
		t.Errorf("Expected saved path %s, got %s", expectedPath, savedPath)
	}

	// List profiles
	profiles, err := cm.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles failed: %v", err)
	}

	if len(profiles) != 1 {
		t.Fatalf("Expected 1 profile, got %d", len(profiles))
	}

	if profiles[0].Name != "test-vpn" {
		t.Errorf("Profile name mismatch: got %s, want test-vpn", profiles[0].Name)
	}
	if profiles[0].ClientIP != "10.8.0.2/32" {
		t.Errorf("Client IP mismatch: got %s, want 10.8.0.2/32", profiles[0].ClientIP)
	}
	if profiles[0].Endpoint != "203.0.113.1:51820" {
		t.Errorf("Endpoint mismatch: got %s, want 203.0.113.1:51820", profiles[0].Endpoint)
	}

	// Delete profile
	if err := cm.DeleteProfile("test-vpn"); err != nil {
		t.Fatalf("DeleteProfile failed: %v", err)
	}

	profilesAfter, _ := cm.ListProfiles()
	if len(profilesAfter) != 0 {
		t.Errorf("Expected 0 profiles after deletion, got %d", len(profilesAfter))
	}
}
