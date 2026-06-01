package agent

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"github.com/gallofrancesco1312/hlab-cli/internal/config"
)

type TLSConfig struct {
	CertFile string
	KeyFile  string
	CAFile   string
}

type Server struct {
	NodeName string
	Version  string
	Port     int
	TLS      TLSConfig
	Mux      *http.ServeMux
}

func NewServer(configFile string) *Server {
	cfg, err := config.AgentLoad()
	if err != nil {
		panic(fmt.Sprintf("failed to load agent config: %v", err))
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
		Mux: http.NewServeMux(),
	}
}

func (s *Server) registerRoutes() {
	s.Mux.HandleFunc("/health", s.handleHealth)
	s.Mux.HandleFunc("/services", s.handleServices)
}

func (s *Server) Start() error {
	if s.TLS.CertFile == "" || s.TLS.KeyFile == "" {
		return fmt.Errorf("TLS certificate and key files must be provided")
	}
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

	addr := fmt.Sprintf(":%d", s.Port)
	server := http.Server{
		Addr:      addr,
		Handler:   s.Mux,
		TLSConfig: tlsCfg,
	}
	return server.ListenAndServeTLS(s.TLS.CertFile, s.TLS.KeyFile)
}
