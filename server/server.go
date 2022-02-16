package server

import (
	"crypto/tls"
	"embed"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"pair-ls/auth"
	"pair-ls/state"
	"strings"
	"text/template"

	_ "embed"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/vearutop/statigz"
)

//go:embed dist/*
var static embed.FS

//go:embed index.html
var index embed.FS
var indexTmpl *template.Template

type WebServerConfig struct {
	// If provided, will require password auth from web client
	WebPassword string `json:"webPassword"`
	// If provided, will require connecting pair-ls LSP to provide this password (only used for relay & signal servers)
	LspPassword string `json:"lspPassword"`
	// If provided, will secure all connections with TLS
	CertFile string `json:"certFile"`
	// If the key is not encoded in the CertFile PEM, you can pass it in separately here
	KeyFile string `json:"keyFile"`
	// If true, will require pair-ls LSP to provide a matching client cert
	RequireClientCert bool `json:"requireClientCert"`
	// PEM file with one or more certs that pair-ls LSP can match (when RequireClientCert = true)
	// (only used for relay & signal servers)
	ClientCAs string `json:"clientCAs"`
}

type WebServer struct {
	logger       *log.Logger
	state        *state.WorkspaceState
	config       WebServerConfig
	relay        *relayServer
	signalServer *signalServer
}

func NewServer(state *state.WorkspaceState, logger *log.Logger, config WebServerConfig) *WebServer {
	return &WebServer{
		logger: logger,
		state:  state,
		config: config,
	}
}

func (s *WebServer) AddRelayServer(handler jsonrpc2.Handler, config RelayConfig) {
	s.relay = &relayServer{
		logger:      s.logger,
		handler:     handler,
		state:       s.state,
		config:      config,
		webConfig:   s.config,
		connCounter: make(chan bool),
	}
}

func (s *WebServer) MakeSignalServer() {
	s.signalServer = &signalServer{
		logger:    s.logger,
		editorMap: make(map[string]*jsonrpc2.Conn),
		webConfig: s.config,
	}
}

func (s *WebServer) Serve(hostname string, port int) {
	if s.config.WebPassword == "" && s.signalServer == nil {
		s.logger.Println("WARNING: running webserver with no password")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/client_ws", s.on_websocket)
	mux.HandleFunc("/login", s.on_login)
	mux.HandleFunc("/", s.on_index)
	mux.Handle("/dist/", statigz.FileServer(static))
	if s.relay != nil {
		s.relay.attachHandlers(mux)
	}
	if s.signalServer != nil {
		s.signalServer.attachHandlers(mux)
	}
	tlsConfig, err := createTLSConfig(s.config)
	if err != nil {
		s.logger.Fatalln(err)
	}
	if tlsConfig != nil && (port == 80 || port == 443) {
		go s.serveHTTPSRedirect(hostname, 80)
		port = 443
	}
	host := fmt.Sprintf("%s:%d", hostname, port)
	s.logger.Printf("Listening for connections on %s\n", host)
	defer s.logger.Println("Server shut down")
	srv := &http.Server{
		Addr:      host,
		Handler:   mux,
		TLSConfig: tlsConfig,
	}
	if tlsConfig != nil {
		log.Fatal(srv.ListenAndServeTLS("", ""))
	} else {
		log.Fatal(srv.ListenAndServe())
	}
}

func (s *WebServer) on_index(w http.ResponseWriter, r *http.Request) {
	if indexTmpl == nil {
		var err error
		indexTmpl, err = template.ParseFS(index, "index.html")
		if err != nil {
			s.logger.Println("Error parsing index.html template", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
	useRTC := s.signalServer != nil
	indexTmpl.Execute(w, useRTC)
}

func createTLSConfig(conf WebServerConfig) (*tls.Config, error) {
	if conf.CertFile == "" {
		return nil, nil
	}
	cert, pool, err := auth.LoadCertFromPEMs(conf.CertFile, conf.KeyFile)
	if err != nil {
		return nil, err
	}
	auth := tls.NoClientCert
	if conf.RequireClientCert {
		auth = tls.VerifyClientCertIfGiven
		if conf.ClientCAs != "" {
			data, err := ioutil.ReadFile(conf.ClientCAs)
			if err != nil {
				return nil, err
			}
			pool.AppendCertsFromPEM(data)
		}
	}

	config := tls.Config{
		ClientAuth:   auth,
		Certificates: []tls.Certificate{*cert},
		ClientCAs:    pool,
	}
	return &config, nil
}

func (c *WebServerConfig) requireAuth(w http.ResponseWriter, r *http.Request) error {
	var err error
	if c.RequireClientCert && len(r.TLS.VerifiedChains) == 0 {
		err = errors.New("No valid client certificate")
	}
	if c.LspPassword != "" {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			err = errors.New("Missing Authorization header")
		} else {
			pieces := strings.SplitN(auth, " ", 2)
			if len(pieces) < 2 {
				err = errors.New("Malformed Authorization header")
			} else {
				var code []byte
				code, err = base64.StdEncoding.DecodeString(pieces[1])
				if string(code) != c.LspPassword {
					err = errors.New("Password mismatch")
				}
			}
		}
	}
	if err != nil {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	}
	return err
}
