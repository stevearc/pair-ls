package auth

import (
	"crypto"
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

func GetTLSDialer(certFile string, keyFile string) (*websocket.Dialer, error) {
	privateKey, err := LoadPrivateKeyPEM(keyFile)
	if err != nil {
		return nil, err
	}
	pool, leaf, err := LoadCertFromPEM(certFile)
	if err != nil {
		return nil, err
	}

	tlsCert := tls.Certificate{
		Certificate: [][]byte{leaf.Raw},
		PrivateKey:  privateKey,
		Leaf:        leaf,
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

func LoadCertFromPEM(filename string) (*x509.CertPool, *x509.Certificate, error) {
	pool := x509.NewCertPool()
	certBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}
	ok := pool.AppendCertsFromPEM(certBytes)
	if !ok {
		return nil, nil, errors.New(fmt.Sprintf("Failed to parse root certificate %s", filename))
	}
	block, _ := pem.Decode(certBytes)
	if block == nil {
		return nil, nil, errors.New(fmt.Sprintf("Failed to parse certificate PEM %s", filename))
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, err
	}
	return pool, cert, nil
}

func LoadPrivateKeyPEM(filename string) (crypto.PrivateKey, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode([]byte(data))
	if block == nil {
		return nil, errors.New(fmt.Sprintf("failed to parse private key PEM: %s", filename))
	}
	return x509.ParsePKCS8PrivateKey(block.Bytes)
}
