package lsp_handler

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"pair-ls/state"
	"sync"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type LspHandler struct {
	logger            *log.Logger
	config            *HandlerConfig
	lspConn           *jsonrpc2.Conn
	rootPath          string
	initialized       bool
	state             *state.WorkspaceState
	clientSendsCursor bool
	changeTextChan    chan TextChange
	forwardChan       chan *jsonrpc2.Request
	peerMap           map[string]*webrtc.PeerConnection
	rtc               *webrtc.API
	mu                sync.Mutex
	pendingNotifs     []pendingNotif
}

type HandlerConfig struct {
	RelayServer   string
	SignalServer  string
	StaticRTCSite string
	ClientAuth    ClientAuthConfig
}

func NewHandler(state *state.WorkspaceState, logger *log.Logger, config *HandlerConfig) *LspHandler {
	changeTextChan := make(chan TextChange)
	s := webrtc.SettingEngine{}
	s.DetachDataChannels()

	handler := &LspHandler{
		logger:         logger,
		config:         config,
		state:          state,
		rtc:            webrtc.NewAPI(webrtc.WithSettingEngine(s)),
		changeTextChan: changeTextChan,
		peerMap:        make(map[string]*webrtc.PeerConnection),
		pendingNotifs:  make([]pendingNotif, 0),
	}
	// TODO: make this configurable
	go debounceChangeText(200*time.Millisecond, handler.changeTextChan, func(change TextChange) {
		handler.state.ReplaceText(change.Filename, change.Text, !handler.clientSendsCursor)
	})

	return handler
}

func (h *LspHandler) GetRPCHandler() jsonrpc2.Handler {
	return jsonrpc2.HandlerWithError(h.handle)
}

func (h *LspHandler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			h.logger.Println("Error handling", req.Method, r)
		}
	}()

	if h.forwardChan != nil {
		h.forwardChan <- req
	}
	switch req.Method {
	case "initialize":
		return h.handleInitialize(ctx, conn, req)
	case "initialized":
		return h.handleInitialized(ctx, conn, req)
	case "shutdown":
		return h.handleShutdown(ctx, conn, req)
	case "textDocument/didOpen":
		return h.handleTextDocumentDidOpen(ctx, conn, req)
	case "textDocument/didChange":
		return h.handleTextDocumentDidChange(ctx, conn, req)
	case "textDocument/didClose":
		return h.handleTextDocumentDidClose(ctx, conn, req)
	case "textDocument/hover":
		return h.handleTextDocumentHover(ctx, conn, req)
	case "experimental/cursor":
		return h.handleCursorMove(ctx, conn, req)
	case "experimental/connectToPeer":
		return h.handleConnectToPeer(ctx, conn, req)
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}

type TextChange struct {
	Filename string
	Text     string
}

func debounceChangeText(interval time.Duration, input chan TextChange, cb func(arg TextChange)) {
	var item TextChange
	running := false
	ticker := time.NewTicker(interval)
	ticker.Stop()
	for {
		select {
		case item = <-input:
			if !running {
				ticker.Reset(interval)
				running = true
			}
		case <-ticker.C:
			cb(item)
			item = TextChange{}
			ticker.Stop()
			running = false
		}
	}
}

func (h *LspHandler) ListenOnStdin(logger *log.Logger, loglevel int, callToken string) {
	if h.config.SignalServer != "" {
		go h.listenForRTC(h.config.SignalServer, h.config.ClientAuth)
	}

	if h.config.RelayServer != "" {
		h.forwardChan = make(chan *jsonrpc2.Request)
		go h.forward(h.config.ClientAuth)
	}

	var connOpt []jsonrpc2.ConnOpt
	if loglevel >= 5 {
		connOpt = append(connOpt, jsonrpc2.LogMessages(logger))
	}

	if callToken != "" {
		answer, err := h.respondRTCPeer(callToken)
		if err != nil {
			h.logger.Println("Failed to respond to WebRTC call", err)
		}
		h.SendShareString(answer)
	}

	logger.Println("Server listening on stdin")
	defer logger.Println("Server stopped")
	h.lspConn = jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(stdrwc{}, jsonrpc2.VSCodeObjectCodec{}),
		h.GetRPCHandler(), connOpt...)
	<-h.lspConn.DisconnectNotify()
}

type AuthRequest struct {
	Token string `json:"token"`
}

func (h *LspHandler) SendShareString(shareValue string) {
	h.showMessage(fmt.Sprintf("Sharing: %s", shareValue), lsp.Info)
}

func (h *LspHandler) showMessage(message string, mType lsp.MessageType) {
	method := "window/showMessage"
	params := lsp.ShowMessageParams{
		Type:    mType,
		Message: message,
	}
	if h.lspConn == nil || !h.initialized {
		h.pendingNotifs = append(h.pendingNotifs, pendingNotif{
			method: method,
			params: params,
		})
	} else {
		err := h.lspConn.Notify(context.Background(), method, params)
		if err != nil {
			h.logger.Println("Error send message to client", message)
		}
	}
}

func (h *LspHandler) createStaticUrl(offer string) string {
	return h.config.StaticRTCSite + "?t=" + url.QueryEscape(offer)
}

type pendingNotif struct {
	method string
	params interface{}
}

type stdrwc struct{}

func (stdrwc) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (c stdrwc) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (c stdrwc) Close() error {
	if err := os.Stdin.Close(); err != nil {
		return err
	}
	return os.Stdout.Close()
}
