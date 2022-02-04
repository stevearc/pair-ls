package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/rakyll/command"
)

type certCommand struct {
	outfile string
	dns     string
}

func NewCertCmd(conf *PairConfig) command.Cmd {
	return &certCommand{}
}

func (cmd *certCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.outfile, "out", "server", "Output file")
	fs.StringVar(&cmd.dns, "dns", "", "Domain name of the host (e.g. www.mycode.com)")
	return fs
}

func (cmd *certCommand) Run(args []string) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	pub := &priv.PublicKey

	dns := []string{"localhost"}
	if cmd.dns != "" {
		dns = append(dns, cmd.dns)
	}
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Country:            []string{"US"},
			Organization:       []string{"PairLS"},
			OrganizationalUnit: []string{""},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(5, 0, 0),
		BasicConstraintsValid: true,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		DNSNames:              dns,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, pub, priv)
	if err != nil {
		log.Println("create ca failed", err)
		return
	}

	ca_file := fmt.Sprintf("%s.pem", cmd.outfile)
	certOut, err := os.Create(ca_file)
	if err != nil {
		log.Fatalf("Failed to open %s for writing: %v\n", ca_file, err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		log.Fatalf("Failed to write data to %s: %v\n", ca_file, err)
	}
	if err := certOut.Close(); err != nil {
		log.Fatalf("Error closing %s: %v\n", ca_file, err)
	}
	fmt.Printf("wrote %s\n", ca_file)

	keyFile := fmt.Sprintf("%s.key.pem", cmd.outfile)
	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Failed to open %s for writing: %v\n", keyFile, err)
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		log.Fatalf("Unable to marshal private key: %v\n", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		log.Fatalf("Failed to write data to %s: %v\n", keyFile, err)
	}
	if err := keyOut.Close(); err != nil {
		log.Fatalf("Error closing %s: %v\n", keyFile, err)
	}
	fmt.Printf("wrote %s\n", keyFile)
}
