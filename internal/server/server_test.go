package server

import (
	"testing"

	"github.com/ezwg/ezwg/internal/wireguard"
)

func TestAllocateNextIP(t *testing.T) {
	cfg := &wireguard.Config{
		Interface: wireguard.InterfaceConfig{
			Address: []string{"10.8.0.1/24"},
		},
		Peers: []wireguard.PeerConfig{
			{
				Name:       "peer1",
				AllowedIPs: []string{"10.8.0.2/32"},
			},
			{
				Name:       "peer2",
				AllowedIPs: []string{"10.8.0.3/32"},
			},
		},
	}

	ip, err := allocateNextIP(cfg)
	if err != nil {
		t.Fatalf("allocateNextIP failed: %v", err)
	}

	if ip.String() != "10.8.0.4" {
		t.Errorf("Expected next IP to be 10.8.0.4, got %s", ip.String())
	}

	// Now add 10.8.0.4 and test 10.8.0.5
	cfg.Peers = append(cfg.Peers, wireguard.PeerConfig{
		Name:       "peer3",
		AllowedIPs: []string{"10.8.0.4/32"},
	})

	ip2, err := allocateNextIP(cfg)
	if err != nil {
		t.Fatalf("allocateNextIP failed: %v", err)
	}

	if ip2.String() != "10.8.0.5" {
		t.Errorf("Expected next IP to be 10.8.0.5, got %s", ip2.String())
	}
}
