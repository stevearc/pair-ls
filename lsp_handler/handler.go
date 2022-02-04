package lsp_handler

import (
	"context"
	"fmt"
	"log"
	"os"
	"pair-ls/server"
	"pair-ls/state"
	"strings"
	"time"

	"github.com/sourcegraph/jsonrpc2"
)

type lspHandler struct {
	logger            *log.Logger
	rootPath          string
	state             *state.WorkspaceState
	clientSendsCursor bool
	changeTextChan    chan TextChange
	forwardChan       chan *jsonrpc2.Request
}

func NewHandler(state *state.WorkspaceState, logger *log.Logger, config *server.ServerConfig, forwardHost string) jsonrpc2.Handler {
	changeTextChan := make(chan TextChange)
	handler := &lspHandler{
		logger:         logger,
		state:          state,
		changeTextChan: changeTextChan,
	}
	// TODO: make this configurable
	go debounceChangeText(200*time.Millisecond, changeTextChan, func(change TextChange) {
		handler.state.ReplaceText(change.Filename, strings.Split(change.Text, "\n"), !handler.clientSendsCursor)
	})

	if forwardHost != "" {
		handler.forwardChan = make(chan *jsonrpc2.Request)
		go forward(forwardHost, logger, handler.forwardChan, config.CertFile, config.KeyFile)
	}
	return jsonrpc2.HandlerWithError(handler.handle)
}

func (h *lspHandler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
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
		return
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

func ListenOnStdin(handler jsonrpc2.Handler, logger *log.Logger, loglevel int) {
	var connOpt []jsonrpc2.ConnOpt
	if loglevel >= 5 {
		connOpt = append(connOpt, jsonrpc2.LogMessages(logger))
	}

	logger.Println("Server listening on stdin")
	defer logger.Println("Server stopped")
	<-jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(stdrwc{}, jsonrpc2.VSCodeObjectCodec{}),
		handler, connOpt...).DisconnectNotify()
}

type AuthRequest struct {
	Token string `json:"token"`
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
