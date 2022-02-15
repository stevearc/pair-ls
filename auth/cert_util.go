package auth

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func GetTLSDialer(certFile string) (*websocket.Dialer, error) {
	tlsCert, pool, err := LoadCertPem(certFile)
	if err != nil {
		return nil, err
	}
	dialer := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{tlsCert},
			RootCAs:      pool,
		},
	}
	return &dialer, nil
}

func LoadCertPem(filename string) (tls.Certificate, *x509.CertPool, error) {
	pool := x509.NewCertPool()
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return tls.Certificate{}, nil, err
	}
	pool.AppendCertsFromPEM(data)

	var certBytes *pem.Block
	var privBytes *pem.Block
	var block *pem.Block
	for len(data) > 0 {
		block, data = pem.Decode(data)
		if block == nil {
			return tls.Certificate{}, nil, errors.New(fmt.Sprintf("failed to parse PEM: %s", filename))
		}
		if block.Type == "CERTIFICATE" {
			certBytes = block
		} else {
			privBytes = block
		}
	}
	cert, err := tls.X509KeyPair(pem.EncodeToMemory(certBytes), pem.EncodeToMemory(privBytes))
	return cert, pool, err
}
