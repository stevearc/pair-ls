package server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pair-ls/state"
	"pair-ls/util"

	_ "embed"

	"github.com/gorilla/websocket"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/vearutop/statigz"
	"golang.org/x/crypto/bcrypt"
)

type ServerConfig struct {
	Password string
	KeyFile  string
	CertFile string
}

type Server struct {
	logger *log.Logger
	state  *state.WorkspaceState
	config ServerConfig
}

func NewServer(state *state.WorkspaceState, logger *log.Logger, config ServerConfig) *Server {
	return &Server{
		logger: logger,
		state:  state,
		config: config,
	}
}

//go:embed dist/* index.html
var static embed.FS

func (s *Server) on_websocket(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{} // use default options
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Print("ws upgrade error:", err)
		return
	}
	defer c.Close()
	s.logger.Println("Client connected")
	defer s.logger.Println("Client disconnected")

	handler := websocketHandler{
		logger:   s.logger,
		state:    s.state,
		password: s.config.Password,
	}

	conn := jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(util.WrapWebsocket(c), jsonrpc2.PlainObjectCodec{}),
		jsonrpc2.HandlerWithError(handler.handle),
	)
	handler.run(conn)
}

func (s *Server) on_login(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token := ""
	if s.config.Password != "" {
		if data.Password == s.config.Password {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(s.config.Password), bcrypt.DefaultCost)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			token = string(hashedPassword)
		} else if err := bcrypt.CompareHashAndPassword([]byte(data.Password), []byte(s.config.Password)); err == nil {
			token = data.Password
		} else {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Token string `json:"token"`
	}{
		Token: token,
	})
}

func (s *Server) Serve(hostname string, port int) {
	if s.config.Password == "" {
		s.logger.Println("WARNING: running webserver with no password")
	}
	server := http.NewServeMux()
	server.HandleFunc("/ws", s.on_websocket)
	server.HandleFunc("/login", s.on_login)
	server.Handle("/", statigz.FileServer(static))
	defer s.logger.Println("Server shut down")
	if s.config.KeyFile != "" && s.config.CertFile != "" {
		if port == 80 || port == 443 {
			go s.serveHTTPSRedirect(hostname, 80)
			port = 443
		}
		host := fmt.Sprintf("%s:%d", hostname, port)
		s.logger.Printf("Listening for https connections on %s\n", host)
		log.Fatal(http.ListenAndServeTLS(host, s.config.CertFile, s.config.KeyFile, server))
	} else {
		host := fmt.Sprintf("%s:%d", hostname, port)
		s.logger.Printf("Listening for http connections on %s\n", host)
		log.Fatal(http.ListenAndServe(host, server))
	}
}

func (s *Server) redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, fmt.Sprintf("https://%s", r.Host), 301)
}

func (s *Server) serveHTTPSRedirect(hostname string, port int) {
	server := http.NewServeMux()
	server.HandleFunc("/", s.redirect)
	host := fmt.Sprintf("%s:%d", hostname, port)
	s.logger.Printf("Serving http->https redirects on %s\n", host)
	defer s.logger.Println("Redirect server shut down")
	log.Fatal(http.ListenAndServe(host, server))
}
