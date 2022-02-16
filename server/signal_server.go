package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pair-ls/util"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/randutil"
	"github.com/pion/webrtc/v3"
	"github.com/sourcegraph/jsonrpc2"
)

type signalServer struct {
	logger    *log.Logger
	editorMap map[string]*jsonrpc2.Conn
	mu        sync.Mutex
	webConfig WebServerConfig
}

func (s *signalServer) attachHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/signal", s.on_websocket)
	mux.HandleFunc("/call", s.on_call)
	mux.HandleFunc("/ice", s.on_ice)
}

func (s *signalServer) on_websocket(w http.ResponseWriter, r *http.Request) {
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
	s.logger.Println("Editor connected")
	defer s.logger.Println("Editor disconnected")

	conn := jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(util.WrapWebsocket(c), jsonrpc2.PlainObjectCodec{}),
		jsonrpc2.HandlerWithError(s.handle),
	)
	token, err := createToken()
	if err != nil {
		s.logger.Println("Error creating token", err)
		return
	}
	s.setConn(token, conn)
	defer s.delConn(token)
	conn.Notify(context.Background(), "register", RegisterResponse{Token: token})
	<-conn.DisconnectNotify()
}

func (s *signalServer) setConn(token string, conn *jsonrpc2.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.editorMap[token] = conn
}

func (s *signalServer) getConn(token string) *jsonrpc2.Conn {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.editorMap[token]
}

func (s *signalServer) delConn(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.editorMap, token)
}

func createToken() (string, error) {
	return randutil.GenerateCryptoRandomString(10, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
}

// JSONRPC server

type RegisterResponse struct {
	Token string `json:"token"`
}

func (s *signalServer) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Println("Error handling RTC call", req.Method, r)
		}
	}()
	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}

// HTTP server

type callParams struct {
	Offer webrtc.SessionDescription `json:"offer"`
	Token string                    `json:"token"`
}
type CallResponse struct {
	Answer   webrtc.SessionDescription `json:"answer"`
	ClientID string                    `json:"client_id"`
}

func (s *signalServer) on_call(w http.ResponseWriter, r *http.Request) {
	var params callParams
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	conn := s.getConn(params.Token)
	if conn == nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var response CallResponse
	err = conn.Call(context.Background(), "call", params.Offer, &response)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type iceParams struct {
	Candidate webrtc.ICECandidateInit `json:"candidate"`
	Token     string                  `json:"token"`
	ClientID  string                  `json:"client_id"`
}

type IceRequest struct {
	Candidate webrtc.ICECandidateInit `json:"candidate"`
	ClientID  string                  `json:"client_id"`
}

func (s *signalServer) on_ice(w http.ResponseWriter, r *http.Request) {
	var params iceParams
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	conn := s.getConn(params.Token)
	if conn == nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = conn.Notify(context.Background(), "ice", IceRequest{
		Candidate: params.Candidate,
		ClientID:  params.ClientID,
	})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct{}{})
}
