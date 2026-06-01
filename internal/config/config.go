// Package config manages the CLI configuration (~/.hlab/config.yaml)
// and the discovered nodes cache (~/.hlab/nodes.json).
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/gallofrancesco1312/hlab-cli/pkg/hlabapi"
)

// Config is the CLI configuration structure (~/.hlab/config.yaml).
// The `yaml:"..."` tags tell the parser how to map YAML keys to Go fields.
type Config struct {
	// Aliases maps short names (e.g. "nas") to real node_names (e.g. "homelab-nas").
	Aliases map[string]string `yaml:"aliases"`

	// StaleAfter is the time after which a node not seen is marked stale.
	// We use time.Duration which deserializes from strings like "15m", "1h".
	StaleAfter time.Duration `yaml:"stale_after"`

	// TLS holds the paths to client certificates.
	TLS TLSConfig `yaml:"tls"`
}

// TLSConfig groups the paths to client TLS files.
type TLSConfig struct {
	CACert     string `yaml:"ca_cert"`
	ClientCert string `yaml:"client_cert"`
	ClientKey  string `yaml:"client_key"`
}

// defaults returns a Config with sensible values ready to use.
// Go has no constructors: a function returning an initialized value is the idiomatic pattern.
func defaults() Config {
	dir := HlabDir()
	return Config{
		Aliases:    map[string]string{},
		StaleAfter: 15 * time.Minute,
		TLS: TLSConfig{
			CACert:     filepath.Join(dir, "ca.crt"),
			ClientCert: filepath.Join(dir, "client.crt"),
			ClientKey:  filepath.Join(dir, "client.key"),
		},
	}
}

// HlabDir returns the path to the ~/.hlab directory, creating it if it does not exist.
func HlabDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Errors are handled explicitly in Go; this is a fatal case
		// because without a home directory the program cannot function.
		panic(fmt.Sprintf("cannot determine home directory: %v", err))
	}
	return filepath.Join(home, ".hlab")
}

// configPath returns the path to the configuration file.
func configPath() string {
	return filepath.Join(HlabDir(), "config.yaml")
}

// Load reads the configuration from disk. If the file does not exist it returns the defaults.
// The signature `(Config, error)` is idiomatic Go: both the value and the error are returned.
func Load() (Config, error) {
	cfg := defaults()

	path := configPath()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// File does not exist yet: that's fine, use the defaults.
		return cfg, nil
	}
	if err != nil {
		return cfg, fmt.Errorf("reading config %s: %w", path, err)
	}

	// yaml.Unmarshal populates the struct fields from the YAML.
	// We pass &cfg (pointer) because Unmarshal needs to modify the value.
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing config %s: %w", path, err)
	}

	return cfg, nil
}

// Save writes the configuration to disk, creating the directory if necessary.
func Save(cfg Config) error {
	dir := HlabDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("serializing config: %w", err)
	}

	return os.WriteFile(configPath(), data, 0600)
}

// ResolveAlias resolves an alias to the real node_name.
// If the alias does not exist, it returns the original name unchanged.
func (c Config) ResolveAlias(nameOrAlias string) string {
	if resolved, ok := c.Aliases[nameOrAlias]; ok {
		return resolved
	}
	return nameOrAlias
}

// --- Node cache ---

// nodesPath returns the path to the node cache file.
func nodesPath() string {
	return filepath.Join(HlabDir(), "nodes.json")
}

// LoadNodes reads the node cache from disk.
// Returns an empty (non-nil) slice if the file does not exist.
func LoadNodes() ([]hlabapi.NodeEntry, error) {
	data, err := os.ReadFile(nodesPath())
	if os.IsNotExist(err) {
		return []hlabapi.NodeEntry{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading nodes cache: %w", err)
	}

	var nodes []hlabapi.NodeEntry
	if err := json.Unmarshal(data, &nodes); err != nil {
		return nil, fmt.Errorf("parsing nodes cache: %w", err)
	}
	return nodes, nil
}

// SaveNodes writes the node cache to disk.
func SaveNodes(nodes []hlabapi.NodeEntry) error {
	if err := os.MkdirAll(HlabDir(), 0700); err != nil {
		return fmt.Errorf("creating hlab directory: %w", err)
	}

	// json.MarshalIndent produces human-readable JSON — useful for manual debugging.
	data, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		return fmt.Errorf("serializing nodes: %w", err)
	}

	return os.WriteFile(nodesPath(), data, 0600)
}

// UpsertNode adds or updates a node in the cache.
// "Upsert" = update-or-insert: a common pattern for duplicate-free caches.
func UpsertNode(nodes []hlabapi.NodeEntry, entry hlabapi.NodeEntry) []hlabapi.NodeEntry {
	for i, n := range nodes {
		if n.Node == entry.Node {
			nodes[i] = entry
			return nodes
		}
	}
	return append(nodes, entry)
}

// MarkStale updates the Stale field of each node based on the last time it was seen.
func MarkStale(nodes []hlabapi.NodeEntry, threshold time.Duration) []hlabapi.NodeEntry {
	now := time.Now()
	for i := range nodes {
		nodes[i].Stale = now.Sub(nodes[i].LastSeen) > threshold
	}
	return nodes
}
