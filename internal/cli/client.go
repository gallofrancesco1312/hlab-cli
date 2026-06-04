package cli

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"github.com/gallofrancesco1312/hlab-cli/internal/config"
	"github.com/gallofrancesco1312/hlab-cli/pkg/hlabapi"
)

// newHTTPClient builds an mTLS-capable HTTP client using certs from cfg.
func newHTTPClient(cfg config.Config) (*http.Client, error) {
	caCert, err := os.ReadFile(cfg.TLS.CACert)
	if err != nil {
		return nil, fmt.Errorf("reading CA cert: %w", err)
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caCert)

	cert, err := tls.LoadX509KeyPair(cfg.TLS.ClientCert, cfg.TLS.ClientKey)
	if err != nil {
		return nil, fmt.Errorf("loading client cert: %w", err)
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
	}
	return &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsCfg},
	}, nil
}

// resolveNode looks up a node by name or alias in the local cache.
func resolveNode(nameOrAlias string, cfg config.Config) (hlabapi.NodeEntry, error) {
	name := cfg.ResolveAlias(nameOrAlias)
	nodes, err := config.LoadNodes()
	if err != nil {
		return hlabapi.NodeEntry{}, fmt.Errorf("loading nodes: %w", err)
	}
	for _, n := range nodes {
		if n.Node == name {
			return n, nil
		}
	}
	return hlabapi.NodeEntry{}, fmt.Errorf("node %q not found in cache (try: hlab discover)", nameOrAlias)
}

// nodeURL builds the full HTTPS URL for a given node and path.
func nodeURL(node hlabapi.NodeEntry, path string) string {
	return fmt.Sprintf("https://%s:%d%s", node.Addr, node.Port, path)
}
