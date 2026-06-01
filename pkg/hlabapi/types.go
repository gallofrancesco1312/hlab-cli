// Package hlabapi contains types shared between the CLI and the agent.
// Sharing types avoids duplicating structs and ensures JSON serialization
// is identical on both sides.
package hlabapi

import "time"

// ServiceType indicates which backend manages a service.
type ServiceType string

const (
	ServiceTypeCompose ServiceType = "compose"
	ServiceTypeDocker  ServiceType = "docker"
	ServiceTypeSystemd ServiceType = "systemd"
)

// ServiceStatus represents the current state of a service.
type ServiceStatus string

const (
	StatusRunning ServiceStatus = "running"
	StatusStopped ServiceStatus = "stopped"
	StatusUnknown ServiceStatus = "unknown"
)

// Service describes a single service managed by the agent.
type Service struct {
	Name   string        `json:"name"`
	Type   ServiceType   `json:"type"`
	Status ServiceStatus `json:"status"`
	Uptime string        `json:"uptime,omitempty"`
}

// ServicesResponse is the response from the GET /services endpoint.
type ServicesResponse struct {
	Node     string    `json:"node"`
	Services []Service `json:"services"`
}

// HealthResponse is the response from the GET /health endpoint.
type HealthResponse struct {
	Node    string `json:"node"`
	Version string `json:"version"`
	OK      bool   `json:"ok"`
}

// BeaconPayload is the UDP payload sent by the agent via multicast.
type BeaconPayload struct {
	Node    string `json:"node"`
	Addr    string `json:"addr"`
	Port    int    `json:"port"`
	Version string `json:"version"`
}

// NodeEntry is how a node is stored in the local cache (~/.hlab/nodes.json).
type NodeEntry struct {
	Node     string    `json:"node"`
	Addr     string    `json:"addr"`
	Port     int       `json:"port"`
	Version  string    `json:"version"`
	LastSeen time.Time `json:"last_seen"`
	Stale    bool      `json:"stale"`
}
