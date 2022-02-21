package lsp_handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"pair-ls/server"
	"pair-ls/util"
	"strings"

	"github.com/pion/randutil"
	"github.com/pion/webrtc/v3"
	"github.com/sourcegraph/jsonrpc2"
)

type ClientAuthConfig struct {
	CertFile string `json:"certFile"`
	KeyFile  string `json:"keyFile"`
	Password string `json:"password"`
}

func (h *LspHandler) listenForRTC(signalServer string, config ClientAuthConfig) {
	if !strings.HasSuffix(signalServer, "/signal") {
		signalServer = signalServer + "/signal"
	}
	h.logger.Println("Connecting to signal server", signalServer)
	u, err := url.Parse(signalServer)
	if err != nil {
		h.logger.Fatal("Invalid signal server:", err)
	}
	c, err := wsDialServer(u.String(), config)
	if err != nil {
		h.logger.Fatal("Websocket dial error:", err)
	}

	conn := jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(util.WrapWebsocket(c), jsonrpc2.PlainObjectCodec{}),
		jsonrpc2.HandlerWithError(h.handleSignalRPC),
	)
	defer c.Close()
	<-conn.DisconnectNotify()
}

func (s *LspHandler) setConn(clientID string, conn *webrtc.PeerConnection) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.peerMap[clientID] = conn
}

func (s *LspHandler) getConn(clientID string) *webrtc.PeerConnection {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.peerMap[clientID]
}

func (s *LspHandler) delConn(clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.peerMap, clientID)
}

func createClientID() (string, error) {
	return randutil.GenerateCryptoRandomString(4, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
}

func (h *LspHandler) handleSignalRPC(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			h.logger.Println("Error handling RTC call", req.Method, r)
		}
	}()
	switch req.Method {
	case "register":
		return h.handleRegister(ctx, conn, req)
	case "call":
		return h.handleCall(ctx, conn, req)
	case "ice":
		return h.handleIce(ctx, conn, req)
	}
	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}

func (h *LspHandler) handleRegister(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params server.RegisterResponse
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	h.SendShareString(util.CreateShareURL(h.config.SignalServer, params.Token))
	return nil, nil
}

func (h *LspHandler) handleCall(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var offer webrtc.SessionDescription
	if err := json.Unmarshal(*req.Params, &offer); err != nil {
		return nil, err
	}

	peerConnection, answer, err := h.connectToPeer(offer)
	if err != nil {
		return nil, err
	}
	clientID, err := createClientID()
	if err != nil {
		return nil, err
	}
	h.setConn(clientID, peerConnection)
	h.runPeerConnection(peerConnection, func() {
		h.delConn(clientID)
	})

	return server.CallResponse{
		Answer:   *answer,
		ClientID: clientID,
	}, nil
}

func (h *LspHandler) handleIce(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params server.IceRequest
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}
	peerConn := h.getConn(params.ClientID)
	if peerConn == nil {
		return nil, errors.New(fmt.Sprintf("No peer found with ID %s", params.ClientID))
	}
	peerConn.AddICECandidate(params.Candidate)
	return nil, nil
}

func (h *LspHandler) handlePeerRPC(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			h.logger.Println("Error handling peer RPC", req.Method, r)
		}
	}()
	switch req.Method {
	case "getText":
		return h.handleGetFile(ctx, conn, req)
	}
	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}

func (h *LspHandler) handleGetFile(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params server.GetFileRequest
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	return h.state.GetFile(params.Filename), nil
}
