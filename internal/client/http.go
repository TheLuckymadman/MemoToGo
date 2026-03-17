package client

import (
	"crypto/tls"
	"net/http"
	"time"
)

func NewHTTPClient(httpTimeout time.Duration) *http.Client {
	tr := http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := http.Client{
		Timeout:   httpTimeout,
		Transport: &tr,
	}
	return &client
}
