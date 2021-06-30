package utils

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"os"
	"time"
)

func NewHTTPTransport(rootCAs *x509.CertPool) *http.Transport {
	t := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:           100,
		IdleConnTimeout:        90 * time.Second,
		TLSHandshakeTimeout:    10 * time.Second,
		ExpectContinueTimeout:  1 * time.Second,
		TLSClientConfig:        nil,
		DisableKeepAlives:      false,
		DisableCompression:     false,
		MaxIdleConnsPerHost:    0,
		MaxConnsPerHost:        0,
		ResponseHeaderTimeout:  0,
		TLSNextProto:           nil,
		ProxyConnectHeader:     nil,
		MaxResponseHeaderBytes: 0,
	}

	if rootCAs != nil {
		tlsConfig := &tls.Config{
			RootCAs: rootCAs,
		}

		skipVerify := os.Getenv("INSECURE_SKIP_VERIFY") == "1"
		if skipVerify {
			tlsConfig.InsecureSkipVerify = true
		}
		t.TLSClientConfig = tlsConfig
	}

	return t
}
