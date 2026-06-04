package agent

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

type DockerClient struct {
	client *client.Client
	ctx    context.Context
}

func setupDockerClient() (*DockerClient, error) {
	ctx := context.Background()
	apiClient, err := client.New(client.FromEnv, client.WithUserAgent("hlab-agent/1.0.0"))
	if err != nil {
		panic(err)
	}
	return &DockerClient{client: apiClient, ctx: ctx}, nil
}

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
		Mux:          http.NewServeMux(),
		DockerClient: dockerClient,
	}
}

func (s *Server) registerRoutes() {
	s.Mux.HandleFunc("/health", s.handleHealth)
	s.Mux.HandleFunc("GET /services", s.handleServices)
	s.Mux.HandleFunc("POST /services/{name}/start", s.handleStartService)
	s.Mux.HandleFunc("POST /services/{name}/stop", s.handleStopService)
}

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
