package server

import (
	"context"
	"log"
	"net/http"
	"pair-ls/state"
	"pair-ls/util"

	"github.com/gorilla/websocket"
	"github.com/sourcegraph/jsonrpc2"
)

type relayServer struct {
	logger      *log.Logger
	handler     jsonrpc2.Handler
	state       *state.WorkspaceState
	webConfig   WebServerConfig
	config      RelayConfig
	connCounter chan bool
	connections int
}

type RelayConfig struct {
	Persist bool
}

func (s *relayServer) attachHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/relay", s.on_websocket)
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
}

func (s *relayServer) on_websocket(w http.ResponseWriter, r *http.Request) {
	err := s.webConfig.requireAuth(w, r)
	if err != nil {
		return
	}
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
