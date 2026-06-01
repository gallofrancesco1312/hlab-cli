package agent

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
)

type TLSConfig struct {
	CertFile string
	KeyFile  string
	CAFile   string
}

type Server struct {
	NodeName string
	Port     int
	TLS      TLSConfig
	Mux      *http.ServeMux
}

func (s *Server) Start() error {
	if s.TLS.CertFile == "" || s.TLS.KeyFile == "" {
		return fmt.Errorf("TLS certificate and key files must be provided")
	}
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
