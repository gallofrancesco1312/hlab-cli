package agent

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gallofrancesco1312/hlab-cli/pkg/hlabapi"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

func uptimeFromUnix(createdAt int64) string {
	d := time.Since(time.Unix(createdAt, 0))
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	case hours > 0:
		return fmt.Sprintf("%dh %dm", hours, minutes)
	default:
		return fmt.Sprintf("%dm", minutes)
	}
}

func containerStateToStatus(state container.ContainerState) hlabapi.ServiceStatus {
	switch state {
	case container.StateRunning:
		return hlabapi.ServiceStatusRunning
	case container.StateExited:
		return hlabapi.ServiceStatusExited
	case container.StatePaused:
		return hlabapi.ServiceStatusPaused
	default:
		return hlabapi.ServiceStatusUnknown
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	resp := hlabapi.HealthResponse{
		Node:    s.NodeName,
		Version: s.Version,
		OK:      true,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleServices(w http.ResponseWriter, r *http.Request) {
	ctx := s.DockerClient.ctx
	apiClient := s.DockerClient.client

	w.Header().Set("Content-Type", "application/json")

	containers, err := apiClient.ContainerList(ctx, client.ContainerListOptions{})
	if err != nil {
		json.NewEncoder(w).Encode(err)
		slog.Error("Failed to list containers", "error", err)
		return
	}

	slog.Info("Found containers", "count", len(containers.Items))
	services := make([]hlabapi.Service, 0, len(containers.Items))
	for _, container := range containers.Items {
		fmt.Println(container.Names)
		services = append(services, hlabapi.Service{
			Name:   container.Names[0],
			Id:     container.ID,
			Type:   hlabapi.ServiceTypeDocker,
			Status: containerStateToStatus(container.State),
			Uptime: uptimeFromUnix(container.Created),
		})
	}

	resp := hlabapi.ServicesResponse{
		Node:     s.NodeName,
		Services: services,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleStartService(w http.ResponseWriter, r *http.Request) {
	serviceToStart := r.PathValue("name")
	w.Header().Set("Content-Type", "application/json")

	if _, err := s.DockerClient.client.ContainerStart(s.DockerClient.ctx, serviceToStart, client.ContainerStartOptions{}); err != nil {
		slog.Error("failed to start container", "name", serviceToStart, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(hlabapi.HealthResponse{Node: s.NodeName, Version: s.Version, OK: false})
		return
	}

	slog.Info("started container", "name", serviceToStart)
	json.NewEncoder(w).Encode(hlabapi.HealthResponse{Node: s.NodeName, Version: s.Version, OK: true})
}

func (s *Server) handleStopService(w http.ResponseWriter, r *http.Request) {
	serviceToStop := r.PathValue("name")
	w.Header().Set("Content-Type", "application/json")

	if _, err := s.DockerClient.client.ContainerStop(s.DockerClient.ctx, serviceToStop, client.ContainerStopOptions{}); err != nil {
		slog.Error("failed to stop container", "name", serviceToStop, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(hlabapi.HealthResponse{Node: s.NodeName, Version: s.Version, OK: false})
		return
	}

	slog.Info("stopped container", "name", serviceToStop)
	json.NewEncoder(w).Encode(hlabapi.HealthResponse{Node: s.NodeName, Version: s.Version, OK: true})
}
