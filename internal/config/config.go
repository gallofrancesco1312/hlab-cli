// Package config gestisce la configurazione della CLI (~/.hlab/config.yaml)
// e la cache dei nodi scoperti (~/.hlab/nodes.json).
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

// Config è la struttura della configurazione CLI (~/.hlab/config.yaml).
// I tag `yaml:"..."` dicono al parser come mappare le chiavi YAML sui campi Go.
type Config struct {
	// Aliases mappa nomi corti (es. "nas") ai node_name reali (es. "homelab-nas").
	Aliases map[string]string `yaml:"aliases"`

	// StaleAfter è il tempo dopo cui un nodo non visto viene marcato stale.
	// Usiamo time.Duration che si deserializza da stringhe come "15m", "1h".
	StaleAfter time.Duration `yaml:"stale_after"`

	// TLS contiene i percorsi ai certificati client.
	TLS TLSConfig `yaml:"tls"`
}

// TLSConfig raggruppa i percorsi ai file TLS del client.
type TLSConfig struct {
	CACert     string `yaml:"ca_cert"`
	ClientCert string `yaml:"client_cert"`
	ClientKey  string `yaml:"client_key"`
}

// defaults restituisce una Config con valori sensati pronti all'uso.
// In Go non esistono costruttori: si usa una funzione che ritorna il valore inizializzato.
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

// HlabDir ritorna il percorso della directory ~/.hlab, creandola se non esiste.
func HlabDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// In Go gli errori si gestiscono esplicitamente; qui è un caso fatale
		// perché senza home directory il programma non può funzionare.
		panic(fmt.Sprintf("impossibile determinare home directory: %v", err))
	}
	return filepath.Join(home, ".hlab")
}

// configPath ritorna il percorso del file di configurazione.
func configPath() string {
	return filepath.Join(HlabDir(), "config.yaml")
}

// Load legge la configurazione da disco. Se il file non esiste ritorna i defaults.
// La firma `(Config, error)` è idiomatica Go: si ritornano sia il valore che l'errore.
func Load() (Config, error) {
	cfg := defaults()

	path := configPath()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// File non esiste ancora: va bene, usiamo i defaults
		return cfg, nil
	}
	if err != nil {
		return cfg, fmt.Errorf("lettura config %s: %w", path, err)
	}

	// yaml.Unmarshal popola i campi della struct partendo dal YAML.
	// Passiamo &cfg (puntatore) perché Unmarshal deve modificare il valore.
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing config %s: %w", path, err)
	}

	return cfg, nil
}

// Save scrive la configurazione su disco, creando la directory se necessario.
func Save(cfg Config) error {
	dir := HlabDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creazione directory %s: %w", dir, err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("serializzazione config: %w", err)
	}

	return os.WriteFile(configPath(), data, 0600)
}

// ResolveAlias risolve un alias al node_name reale.
// Se l'alias non esiste, ritorna il nome originale invariato.
func (c Config) ResolveAlias(nameOrAlias string) string {
	if resolved, ok := c.Aliases[nameOrAlias]; ok {
		return resolved
	}
	return nameOrAlias
}

// --- Cache nodi ---

// nodesPath ritorna il percorso della cache dei nodi.
func nodesPath() string {
	return filepath.Join(HlabDir(), "nodes.json")
}

// LoadNodes legge la cache dei nodi da disco.
// Ritorna una slice vuota (non nil) se il file non esiste.
func LoadNodes() ([]hlabapi.NodeEntry, error) {
	data, err := os.ReadFile(nodesPath())
	if os.IsNotExist(err) {
		return []hlabapi.NodeEntry{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("lettura nodes cache: %w", err)
	}

	var nodes []hlabapi.NodeEntry
	if err := json.Unmarshal(data, &nodes); err != nil {
		return nil, fmt.Errorf("parsing nodes cache: %w", err)
	}
	return nodes, nil
}

// SaveNodes scrive la cache dei nodi su disco.
func SaveNodes(nodes []hlabapi.NodeEntry) error {
	if err := os.MkdirAll(HlabDir(), 0700); err != nil {
		return fmt.Errorf("creazione directory hlab: %w", err)
	}

	// json.MarshalIndent produce JSON leggibile — utile per debug manuale del file.
	data, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		return fmt.Errorf("serializzazione nodes: %w", err)
	}

	return os.WriteFile(nodesPath(), data, 0600)
}

// UpsertNode aggiunge o aggiorna un nodo nella cache.
// "Upsert" = update-or-insert: pattern comune per cache senza duplicati.
func UpsertNode(nodes []hlabapi.NodeEntry, entry hlabapi.NodeEntry) []hlabapi.NodeEntry {
	for i, n := range nodes {
		if n.Node == entry.Node {
			nodes[i] = entry
			return nodes
		}
	}
	return append(nodes, entry)
}

// MarkStale aggiorna il campo Stale di ogni nodo in base all'ultima volta che è stato visto.
func MarkStale(nodes []hlabapi.NodeEntry, threshold time.Duration) []hlabapi.NodeEntry {
	now := time.Now()
	for i := range nodes {
		nodes[i].Stale = now.Sub(nodes[i].LastSeen) > threshold
	}
	return nodes
}
