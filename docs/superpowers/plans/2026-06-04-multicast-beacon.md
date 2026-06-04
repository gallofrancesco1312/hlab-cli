# Multicast Beacon Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add UDP multicast beacon broadcasting to hlab-agent, with interval and address sourced from the agent config file.

**Architecture:** `BeaconConfig` is added to `AgentConfig` and parsed from YAML. A `Beacon` struct in a new `internal/agent/beacon.go` sends `BeaconPayload` as JSON over UDP multicast on a ticker. `Server.Start()` gains graceful shutdown via `signal.NotifyContext` and launches the beacon goroutine if enabled.

**Tech Stack:** Go standard library — `net` (UDP), `encoding/json`, `os/signal`, `syscall`, `time`. No new dependencies.

---

## File Map

| File | Action | Responsibility |
|------|--------|---------------|
| `internal/agent_config/agent_config.go` | Modify | Add `BeaconConfig` struct, extend `AgentConfig`, update `AgentDefaults` |
| `internal/agent_config/agent_config_test.go` | Create | Test YAML parsing of beacon section |
| `internal/agent/beacon.go` | Create | `Beacon` struct, `NewBeacon`, `Start`, `send`, `detectLocalIP` |
| `internal/agent/beacon_test.go` | Create | Test `detectLocalIP` and `send` |
| `internal/agent/server.go` | Modify | Add `BeaconCfg` to `Server`, wire `NewBeacon` + `Start`, add graceful shutdown |

---

### Task 1: Add BeaconConfig to AgentConfig

**Files:**
- Modify: `internal/agent_config/agent_config.go`
- Create: `internal/agent_config/agent_config_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/agent_config/agent_config_test.go`:

```go
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
```

- [ ] **Step 2: Run the test to confirm it fails**

```bash
cd /home/piccolo-lord/public-repos/hlab-cli
go test ./internal/agent_config/... -v
```

Expected: compilation error — `AgentConfig` has no field `Beacon`.

- [ ] **Step 3: Implement BeaconConfig**

In `internal/agent_config/agent_config.go`, add the struct and extend `AgentConfig` and `AgentDefaults`:

```go
import (
    "fmt"
    "os"
    "path/filepath"
    "time"                    // add this

    "gopkg.in/yaml.v3"
)

// BeaconConfig controls multicast beacon broadcasting.
type BeaconConfig struct {
    Enabled       bool          `yaml:"enabled"`
    Interval      time.Duration `yaml:"interval"`
    MulticastAddr string        `yaml:"multicast_addr"`
}

type AgentConfig struct {
    NodeName string         `yaml:"node_name"`
    Port     int            `yaml:"port"`
    TLS      AgentTLSConfig `yaml:"tls"`
    Beacon   BeaconConfig   `yaml:"beacon"`
}
```

Update `AgentDefaults` to include beacon defaults:

```go
func AgentDefaults(configFilename string) AgentConfig {
    dir := HlabAgentDir(configFilename)
    return AgentConfig{
        NodeName: "hlab-node",
        Port:     8443,
        TLS: AgentTLSConfig{
            CACert:     filepath.Join(dir, "ca.crt"),
            ClientCert: filepath.Join(dir, "server.crt"),
            ClientKey:  filepath.Join(dir, "server.key"),
        },
        Beacon: BeaconConfig{
            Enabled:       false,
            Interval:      10 * time.Second,
            MulticastAddr: "239.255.42.42:9999",
        },
    }
}
```

- [ ] **Step 4: Run the tests to confirm they pass**

```bash
go test ./internal/agent_config/... -v
```

Expected output:
```
--- PASS: TestBeaconConfigParsed (0.00s)
--- PASS: TestBeaconConfigDefaults (0.00s)
PASS
```

- [ ] **Step 5: Commit**

```bash
git add internal/agent_config/agent_config.go internal/agent_config/agent_config_test.go
git commit -m "feat: add BeaconConfig to AgentConfig with YAML parsing and defaults"
```

---

### Task 2: Create beacon.go

**Files:**
- Create: `internal/agent/beacon.go`
- Create: `internal/agent/beacon_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/agent/beacon_test.go`:

```go
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
	// Listen on a random UDP port on localhost (not multicast — tests the send mechanics).
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
```

- [ ] **Step 2: Run the tests to confirm they fail**

```bash
go test ./internal/agent/... -v -run "TestDetectLocalIP|TestBeaconSend"
```

Expected: compilation error — `Beacon`, `detectLocalIP` undefined.

- [ ] **Step 3: Implement beacon.go**

Create `internal/agent/beacon.go`:

```go
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
```

- [ ] **Step 4: Run the tests to confirm they pass**

```bash
go test ./internal/agent/... -v -run "TestDetectLocalIP|TestBeaconSend"
```

Expected output:
```
--- PASS: TestDetectLocalIP (0.00s)
--- PASS: TestBeaconSend (0.00s)
PASS
```

- [ ] **Step 5: Commit**

```bash
git add internal/agent/beacon.go internal/agent/beacon_test.go
git commit -m "feat: implement multicast beacon sender with auto IP detection"
```

---

### Task 3: Wire beacon into server.go

**Files:**
- Modify: `internal/agent/server.go`

- [ ] **Step 1: Update Server struct and NewServer**

Replace the `Server` struct and `NewServer` in `internal/agent/server.go` with:

```go
import (
    "context"
    "crypto/tls"
    "crypto/x509"
    "fmt"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"

    "github.com/gallofrancesco1312/hlab-cli/internal/agent_config"
    "github.com/moby/moby/client"
)

type TLSConfig struct {
    CertFile string
    KeyFile  string
    CAFile   string
}

type Server struct {
    NodeName     string
    Version      string
    Port         int
    TLS          TLSConfig
    BeaconCfg    agent_config.BeaconConfig
    Mux          *http.ServeMux
    DockerClient *DockerClient
}

func NewServer(configFile string) *Server {
    slog.Info("Loading agent configuration...")
    cfg, err := agent_config.AgentLoad(configFile)
    if err != nil {
        slog.Error("Failed to load agent configuration", "error", err)
        panic(fmt.Sprintf("failed to load agent config: %v", err))
    }
    dockerClient, err := setupDockerClient()
    if err != nil {
        slog.Error("Failed to set up Docker client", "error", err)
        panic(fmt.Sprintf("failed to set up Docker client: %v", err))
    }

    return &Server{
        NodeName: cfg.NodeName,
        Version:  "1.0.0",
        Port:     cfg.Port,
        TLS: TLSConfig{
            CertFile: cfg.TLS.ClientCert,
            KeyFile:  cfg.TLS.ClientKey,
            CAFile:   cfg.TLS.CACert,
        },
        BeaconCfg:    cfg.Beacon,
        Mux:          http.NewServeMux(),
        DockerClient: dockerClient,
    }
}
```

- [ ] **Step 2: Update Start() with graceful shutdown and beacon**

Replace the `Start()` method body:

```go
func (s *Server) Start() error {
    slog.Info("Starting hlab-agent server...", "node_name", s.NodeName, "port", s.Port)
    if s.TLS.CertFile == "" || s.TLS.KeyFile == "" {
        return fmt.Errorf("TLS certificate and key files must be provided")
    }

    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    slog.Info("Registering HTTP routes...")
    s.registerRoutes()

    caCert, err := os.ReadFile(s.TLS.CAFile)
    if err != nil {
        return fmt.Errorf("failed to read CA certificate: %w", err)
    }
    pool := x509.NewCertPool()
    pool.AppendCertsFromPEM(caCert)

    tlsCfg := &tls.Config{
        ClientAuth: tls.RequireAndVerifyClientCert,
        ClientCAs:  pool,
    }

    httpServer := &http.Server{
        Addr:      fmt.Sprintf(":%d", s.Port),
        Handler:   s.Mux,
        TLSConfig: tlsCfg,
    }

    if s.BeaconCfg.Enabled {
        beacon, err := NewBeacon(s.BeaconCfg, s.NodeName, s.Port, s.Version)
        if err != nil {
            return fmt.Errorf("creating beacon: %w", err)
        }
        go beacon.Start(ctx)
    }

    errCh := make(chan error, 1)
    go func() {
        slog.Info("Agent server is listening", "port", s.Port)
        errCh <- httpServer.ListenAndServeTLS(s.TLS.CertFile, s.TLS.KeyFile)
    }()

    select {
    case err := <-errCh:
        return err
    case <-ctx.Done():
        slog.Info("Shutdown signal received, stopping server...")
        return httpServer.Shutdown(context.Background())
    }
}
```

- [ ] **Step 3: Build to verify no compilation errors**

```bash
go build ./...
```

Expected: no output (clean build).

- [ ] **Step 4: Manual smoke test**

With the agent running (`./hlab-agent agent.yaml`, which has `beacon.enabled: true`), listen for UDP packets on the multicast group:

```bash
# In a second terminal — join the multicast group and print received packets
socat UDP4-RECVFROM:9999,ip-add-membership=239.255.42.42:0.0.0.0,fork -
```

Expected every 10 seconds: a line like:
```json
{"node":"potato-server","addr":"192.168.x.x","port":8444,"version":"1.0.0"}
```

- [ ] **Step 5: Commit**

```bash
git add internal/agent/server.go
git commit -m "feat: wire multicast beacon into agent server with graceful shutdown"
```
