package agent

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/gallofrancesco1312/hlab-cli/pkg/hlabapi"
)

func TestDetectLocalIP(t *testing.T) {
	ip, err := detectLocalIP()
	if err != nil {
		t.Fatalf("detectLocalIP: %v", err)
	}
	if ip == "" {
		t.Fatal("detectLocalIP returned empty string")
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		t.Fatalf("detectLocalIP returned non-IP string: %q", ip)
	}
	if parsed.IsLoopback() {
		t.Errorf("detectLocalIP returned loopback address: %q", ip)
	}
}

func TestBeaconSend(t *testing.T) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer pc.Close()

	b := &Beacon{
		payload: hlabapi.BeaconPayload{
			Node:    "test-node",
			Addr:    "192.168.1.1",
			Port:    8443,
			Version: "1.0.0",
		},
		addr:     pc.LocalAddr().String(),
		interval: time.Second,
	}

	if err := b.send(); err != nil {
		t.Fatalf("send: %v", err)
	}

	buf := make([]byte, 1024)
	if err := pc.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("SetReadDeadline: %v", err)
	}
	n, _, err := pc.ReadFrom(buf)
	if err != nil {
		t.Fatalf("ReadFrom: %v", err)
	}

	var got hlabapi.BeaconPayload
	if err := json.Unmarshal(buf[:n], &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Node != "test-node" {
		t.Errorf("expected Node=test-node, got %q", got.Node)
	}
	if got.Port != 8443 {
		t.Errorf("expected Port=8443, got %d", got.Port)
	}
}
