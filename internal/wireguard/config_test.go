package wireguard

import (
	"strings"
	"testing"
)

func TestConfigSerializeAndParse(t *testing.T) {
	orig := &Config{
		Interface: InterfaceConfig{
			PrivateKey: "YWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWE=",
			Address:    []string{"10.8.0.1/24"},
			ListenPort: 51820,
			DNS:        []string{"1.1.1.1", "8.8.8.8"},
			MTU:        1420,
			PostUp:     []string{"iptables -A FORWARD -i %i -j ACCEPT"},
			PostDown:   []string{"iptables -D FORWARD -i %i -j ACCEPT"},
		},
		Peers: []PeerConfig{
			{
				Name:                "phone",
				PublicKey:           "YmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmI=",
				PresharedKey:        "Y2NjY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2NjY2M=",
				AllowedIPs:          []string{"10.8.0.2/32"},
				Endpoint:            "203.0.113.10:51820",
				PersistentKeepalive: 25,
			},
		},
	}

	serialized := orig.Serialize()
	parsed, err := ParseConfig(strings.NewReader(serialized))
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	if parsed.Interface.PrivateKey != orig.Interface.PrivateKey {
		t.Errorf("PrivateKey mismatch: got %s, want %s", parsed.Interface.PrivateKey, orig.Interface.PrivateKey)
	}
	if parsed.Interface.ListenPort != orig.Interface.ListenPort {
		t.Errorf("ListenPort mismatch: got %d, want %d", parsed.Interface.ListenPort, orig.Interface.ListenPort)
	}
	if len(parsed.Peers) != 1 {
		t.Fatalf("Expected 1 peer, got %d", len(parsed.Peers))
	}

	peer := parsed.Peers[0]
	if peer.Name != "phone" {
		t.Errorf("Peer name mismatch: got '%s', want 'phone'", peer.Name)
	}
	if peer.PublicKey != orig.Peers[0].PublicKey {
		t.Errorf("Peer public key mismatch: got %s, want %s", peer.PublicKey, orig.Peers[0].PublicKey)
	}
	if peer.PersistentKeepalive != 25 {
		t.Errorf("Peer keepalive mismatch: got %d, want 25", peer.PersistentKeepalive)
	}
}
