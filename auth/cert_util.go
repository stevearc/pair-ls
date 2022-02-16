package auth

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
)

func LoadTLSConfig(filenames ...string) (*tls.Config, error) {
	tlsCert, pool, err := LoadCertFromPEMs(filenames...)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{*tlsCert},
		RootCAs:      pool,
	}, nil
}

func LoadCertFromPEMs(filenames ...string) (*tls.Certificate, *x509.CertPool, error) {
	pool := x509.NewCertPool()
	var certBytes *pem.Block
	var privBytes *pem.Block
	for _, filename := range filenames {
		if filename == "" {
			continue
		}
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, nil, err
		}
		pool.AppendCertsFromPEM(data)

		var block *pem.Block
		for len(data) > 0 {
			block, data = pem.Decode(data)
			if block == nil {
				return nil, nil, errors.New(fmt.Sprintf("failed to parse PEM: %s", filename))
			}
			if block.Type == "CERTIFICATE" {
				if certBytes != nil {
					return nil, nil, errors.New(fmt.Sprintf("Multiple certificates found in file(s) %s", filenames))
				}
				certBytes = block
			} else {
				if privBytes != nil {
					return nil, nil, errors.New(fmt.Sprintf("Multiple keys found in file(s) %s", filenames))
				}
				privBytes = block
			}
		}
	}
	cert, err := tls.X509KeyPair(pem.EncodeToMemory(certBytes), pem.EncodeToMemory(privBytes))
	if err != nil {
		return nil, nil, err
	}
	return &cert, pool, nil
}
