package agent_config

import (
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestBeaconConfigParsed(t *testing.T) {
	input := `
node_name: my-node
port: 8443
beacon:
  enabled: true
  interval: 5s
  multicast_addr: "239.255.42.42:9999"
`
	var cfg AgentConfig
	if err := yaml.Unmarshal([]byte(input), &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !cfg.Beacon.Enabled {
		t.Errorf("expected Beacon.Enabled=true, got false")
	}
	if cfg.Beacon.Interval != 5*time.Second {
		t.Errorf("expected Interval=5s, got %v", cfg.Beacon.Interval)
	}
	if cfg.Beacon.MulticastAddr != "239.255.42.42:9999" {
		t.Errorf("expected MulticastAddr=239.255.42.42:9999, got %q", cfg.Beacon.MulticastAddr)
	}
}

func TestBeaconConfigDefaults(t *testing.T) {
	cfg := AgentDefaults("agent.yaml")
	if cfg.Beacon.Enabled {
		t.Errorf("expected Beacon.Enabled=false by default")
	}
	if cfg.Beacon.Interval != 10*time.Second {
		t.Errorf("expected default Interval=10s, got %v", cfg.Beacon.Interval)
	}
	if cfg.Beacon.MulticastAddr != "239.255.42.42:9999" {
		t.Errorf("expected default MulticastAddr=239.255.42.42:9999, got %q", cfg.Beacon.MulticastAddr)
	}
}
