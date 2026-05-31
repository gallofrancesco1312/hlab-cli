// Package hlabapi contiene i tipi condivisi tra CLI e agente.
// Usare tipi condivisi evita di duplicare struct e garantisce che
// la serializzazione JSON sia identica su entrambi i lati.
package hlabapi

import "time"

// ServiceType indica quale backend gestisce un servizio.
type ServiceType string

const (
	ServiceTypeCompose  ServiceType = "compose"
	ServiceTypeDocker   ServiceType = "docker"
	ServiceTypeSystemd  ServiceType = "systemd"
)

// ServiceStatus rappresenta lo stato corrente di un servizio.
type ServiceStatus string

const (
	StatusRunning  ServiceStatus = "running"
	StatusStopped  ServiceStatus = "stopped"
	StatusUnknown  ServiceStatus = "unknown"
)

// Service descrive un singolo servizio gestito dall'agente.
type Service struct {
	Name   string        `json:"name"`
	Type   ServiceType   `json:"type"`
	Status ServiceStatus `json:"status"`
	Uptime string        `json:"uptime,omitempty"`
}

// ServicesResponse è la risposta del endpoint GET /services.
type ServicesResponse struct {
	Node     string    `json:"node"`
	Services []Service `json:"services"`
}

// HealthResponse è la risposta del endpoint GET /health.
type HealthResponse struct {
	Node    string `json:"node"`
	Version string `json:"version"`
	OK      bool   `json:"ok"`
}

// BeaconPayload è il payload UDP inviato dall'agente via multicast.
type BeaconPayload struct {
	Node    string `json:"node"`
	Addr    string `json:"addr"`
	Port    int    `json:"port"`
	Version string `json:"version"`
}

// NodeEntry è come un nodo viene salvato nella cache locale (~/.hlab/nodes.json).
type NodeEntry struct {
	Node      string    `json:"node"`
	Addr      string    `json:"addr"`
	Port      int       `json:"port"`
	Version   string    `json:"version"`
	LastSeen  time.Time `json:"last_seen"`
	Stale     bool      `json:"stale"`
}
