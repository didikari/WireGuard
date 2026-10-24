package wireguard

import (
	"testing"
)

func TestParseWgDump(t *testing.T) {
	// Sample dump output from wg show wg0 dump:
	// row 1: <privkey>\t<pubkey>\t<listen-port>\t<fwmark>
	// row 2: <pubkey>\t<psk>\t<endpoint>\t<allowed-ips>\t<latest-handshake>\t<rx>\t<tx>\t<keepalive>
	sampleDump := "privateKeyBase64=\tserverPubKeyBase64=\t51820\toff\n" +
		"peerPubKey1=\tpskKey1=\t198.51.100.1:54321\t10.8.0.2/32\t1698100000\t1048576\t2097152\t25\n"

	status, err := ParseWgDump("wg0", sampleDump)
	if err != nil {
		t.Fatalf("ParseWgDump failed: %v", err)
	}

	if status.Name != "wg0" {
		t.Errorf("Expected iface wg0, got %s", status.Name)
	}
	if status.ListenPort != 51820 {
		t.Errorf("Expected port 51820, got %d", status.ListenPort)
	}
	if len(status.Peers) != 1 {
		t.Fatalf("Expected 1 peer, got %d", len(status.Peers))
	}

	peer, ok := status.Peers["peerPubKey1="]
	if !ok {
		t.Fatalf("Peer 'peerPubKey1=' not found in map")
	}

	if peer.Endpoint != "198.51.100.1:54321" {
		t.Errorf("Endpoint mismatch: got %s", peer.Endpoint)
	}
	if peer.TransferRx != 1048576 {
		t.Errorf("TransferRx mismatch: got %d", peer.TransferRx)
	}
	if peer.TransferTx != 2097152 {
		t.Errorf("TransferTx mismatch: got %d", peer.TransferTx)
	}
	if peer.LatestHandshake.Unix() != 1698100000 {
		t.Errorf("LatestHandshake mismatch: got %v", peer.LatestHandshake)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    uint64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.00 KiB"},
		{1048576, "1.00 MiB"},
		{1073741824, "1.00 GiB"},
	}

	for _, tt := range tests {
		got := FormatBytes(tt.bytes)
		if got != tt.expected {
			t.Errorf("FormatBytes(%d) = %s, want %s", tt.bytes, got, tt.expected)
		}
	}
}
