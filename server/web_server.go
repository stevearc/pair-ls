package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pair-ls/util"

	"github.com/gorilla/websocket"
	"github.com/sourcegraph/jsonrpc2"
	"golang.org/x/crypto/bcrypt"
)

func (s *WebServer) on_websocket(w http.ResponseWriter, r *http.Request) {
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
		password: s.config.WebPassword,
	}

	conn := jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(util.WrapWebsocket(c), jsonrpc2.PlainObjectCodec{}),
		jsonrpc2.HandlerWithError(handler.handle),
	)
	handler.run(conn)
}

func (s *WebServer) on_login(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	token := ""
	if s.config.WebPassword != "" {
		if data.Password == s.config.WebPassword {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(s.config.WebPassword), bcrypt.DefaultCost)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			token = string(hashedPassword)
		} else if err := bcrypt.CompareHashAndPassword([]byte(data.Password), []byte(s.config.WebPassword)); err == nil {
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

func redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, fmt.Sprintf("https://%s", r.Host), 301)
}

func (s *WebServer) serveHTTPSRedirect(hostname string, port int) {
	server := http.NewServeMux()
	server.HandleFunc("/", redirect)
	host := fmt.Sprintf("%s:%d", hostname, port)
	s.logger.Printf("Serving http->https redirects on %s\n", host)
	defer s.logger.Println("Redirect server shut down")
	log.Fatal(http.ListenAndServe(host, server))
}
