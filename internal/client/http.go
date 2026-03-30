package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"
)

func NewHTTPClient(httpTimeout time.Duration, certPath string) (*http.Client, error) {
	caCert, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("read certificate: %w", err)
	}

	caCertPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("system certificate pool: %w", err)
	}

	if caCertPool == nil {
		caCertPool = x509.NewCertPool()
	}

	ok := caCertPool.AppendCertsFromPEM(caCert)
	if !ok {
		return nil, fmt.Errorf("failed to append cert")
	}

	tr := http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
	}
	client := http.Client{
		Timeout:   httpTimeout,
		Transport: &tr,
	}
	return &client, nil
}
