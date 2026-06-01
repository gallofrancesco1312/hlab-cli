package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gallofrancesco1312/hlab-cli/pkg/hlabapi"

	"github.com/moby/moby/client"
)

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
	ctx := context.Background()
	apiClient, err := client.New(client.FromEnv, client.WithUserAgent("hlab-agent/1.0.0"))
	if err != nil {
		panic(err)
	}
	defer apiClient.Close()

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
			Type:   hlabapi.ServiceTypeDocker,
			Status: container.State,
			Uptime: strconv.FormatInt(container.Created, 10), // TODO: convert to human-readable uptime
		})
	}

	resp := hlabapi.ServicesResponse{
		Node:     s.NodeName,
		Services: services,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
