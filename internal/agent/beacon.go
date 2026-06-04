package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/gallofrancesco1312/hlab-cli/internal/agent_config"
	"github.com/gallofrancesco1312/hlab-cli/pkg/hlabapi"
)

// Beacon periodically broadcasts a BeaconPayload as JSON over UDP multicast.
type Beacon struct {
	payload  hlabapi.BeaconPayload
	addr     string
	interval time.Duration
}

// NewBeacon creates a Beacon. It auto-detects the local IP to advertise.
func NewBeacon(cfg agent_config.BeaconConfig, nodeName string, port int, version string) (*Beacon, error) {
	ip, err := detectLocalIP()
	if err != nil {
		return nil, fmt.Errorf("detecting local IP: %w", err)
	}
	return &Beacon{
		payload: hlabapi.BeaconPayload{
			Node:    nodeName,
			Addr:    ip,
			Port:    port,
			Version: version,
		},
		addr:     cfg.MulticastAddr,
		interval: cfg.Interval,
	}, nil
}

// Start sends beacons on the configured interval until ctx is cancelled.
// Intended to run as a goroutine.
func (b *Beacon) Start(ctx context.Context) {
	slog.Info("Beacon started", "addr", b.addr, "interval", b.interval)
	ticker := time.NewTicker(b.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			slog.Info("Beacon stopped")
			return
		case <-ticker.C:
			if err := b.send(); err != nil {
				slog.Warn("Beacon send failed", "error", err)
			}
		}
	}
}

// send opens a UDP connection, writes the payload as JSON, and closes.
func (b *Beacon) send() error {
	conn, err := net.Dial("udp", b.addr)
	if err != nil {
		return fmt.Errorf("dialing %s: %w", b.addr, err)
	}
	defer conn.Close()

	data, err := json.Marshal(b.payload)
	if err != nil {
		return fmt.Errorf("marshaling payload: %w", err)
	}

	_, err = conn.Write(data)
	return err
}

// detectLocalIP returns the first non-loopback IPv4 address found on the host.
func detectLocalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("listing interfaces: %w", err)
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if ip4 := ip.To4(); ip4 != nil {
				return ip4.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no non-loopback IPv4 address found")
}
