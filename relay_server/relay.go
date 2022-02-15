package relay_server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"pair-ls/auth"
	"pair-ls/state"
	"pair-ls/util"

	"github.com/gorilla/websocket"
	"github.com/sourcegraph/jsonrpc2"
)

type RelayServer struct {
	logger      *log.Logger
	handler     jsonrpc2.Handler
	state       *state.WorkspaceState
	config      ServerConfig
	connCounter chan bool
	connections int
}

type ServerConfig struct {
	CertFile string
	Persist  bool
}

func NewServer(handler jsonrpc2.Handler, logger *log.Logger, state *state.WorkspaceState, config ServerConfig) *RelayServer {
	return &RelayServer{
		logger:      logger,
		handler:     handler,
		state:       state,
		config:      config,
		connCounter: make(chan bool),
	}
}

func (s *RelayServer) Serve(hostname string, port int) {
	if port <= 0 {
		s.logger.Fatalln("Relay server must specify a port with -relay-port")
	}
	if s.config.CertFile == "" {
		msg := "Relay server requires a certFile"
		s.logger.Println(msg)
		log.Fatalln(msg)
	}
	cert, pool, err := auth.LoadCertPem(s.config.CertFile)
	if err != nil {
		log.Fatalln(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.on_websocket)
	config := tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    pool,
	}
	host := fmt.Sprintf("%s:%d", hostname, port)
	s.logger.Printf("Listening for wss connections on %s\n", host)
	defer s.logger.Println("Server shut down")
	srv := &http.Server{
		Addr:         host,
		Handler:      mux,
		TLSConfig:    &config,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}
	go func() {
		for {
			connected := <-s.connCounter
			if connected {
				s.connections++
			} else {
				s.connections--
				if s.connections == 0 && !s.config.Persist {
					s.state.Clear()
				}
			}
		}
	}()
	log.Fatal(srv.ListenAndServeTLS("", ""))
}

func (s *RelayServer) on_websocket(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{} // use default options
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Print("ws upgrade error:", err)
		return
	}
	defer c.Close()
	s.logger.Println("Forwarding client connected")
	defer s.logger.Println("Forwarding client disconnected")

	s.connCounter <- true
	defer func() { s.connCounter <- false }()
	conn := jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(util.WrapWebsocket(c), jsonrpc2.PlainObjectCodec{}),
		s.handler,
	)
	<-conn.DisconnectNotify()
}
