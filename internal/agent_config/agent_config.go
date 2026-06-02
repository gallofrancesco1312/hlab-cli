// Package agent_config manages the agent configuration (~/.hlab/config.yaml)
// and the discovered nodes cache (~/.hlab/nodes.json).
package agent_config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// AgentConfig is the agent configuration structure (~/.hlab/config.yaml).
// The `yaml:"..."` tags tell the parser how to map YAML keys to Go fields.
type AgentConfig struct {
	// NodeName is the name of the node.
	NodeName string `yaml:"node_name"`
	Port     int    `yaml:"port"`

	// TLS holds the paths to client certificates.
	TLS AgentTLSConfig `yaml:"tls"`
}

// TLSConfig groups the paths to client TLS files.
type AgentTLSConfig struct {
	CACert     string `yaml:"ca_cert"`
	ClientCert string `yaml:"server_cert"`
	ClientKey  string `yaml:"server_key"`
}

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
	}
}

// HlabDir returns the path to the ~/.hlab directory, creating it if it does not exist.
func HlabAgentDir(configFilename string) string {
	return filepath.Join(".", configFilename)
	//return filepath.Join("/etc/hlab", configFilename)
}

// Load reads the configuration from disk. If the file does not exist it returns the defaults.
// The signature `(Config, error)` is idiomatic Go: both the value and the error are returned.
func AgentLoad(configFilename string) (AgentConfig, error) {
	cfg := AgentDefaults(configFilename)

	path := HlabAgentDir(configFilename)
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

/*
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
*/
