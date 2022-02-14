package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"pair-ls/state"

	"github.com/sourcegraph/jsonrpc2"
	"golang.org/x/crypto/bcrypt"
)

type websocketHandler struct {
	logger   *log.Logger
	state    *state.WorkspaceState
	password string
	authed   bool
}

func (h *websocketHandler) handleGetFile(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	var params struct {
		Filename string `json:"filename"`
	}
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}
	text := h.state.GetText(params.Filename)
	return text, nil
}

func (h *websocketHandler) handleAuth(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if h.authed {
		return nil, nil
	}
	var params struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}
	if h.password != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(params.Token), []byte(h.password)); err != nil {
			return nil, &jsonrpc2.Error{Code: 401, Message: "Invalid auth token"}
		}
	}

	h.authed = true
	conn.Notify(context.Background(), "initialize", struct {
		View  *state.View           `json:"view"`
		Files map[string]state.File `json:"files"`
	}{
		View:  h.state.GetView(),
		Files: h.state.GetFiles(),
	})
	return nil, nil
}

func (h *websocketHandler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			h.logger.Println("Error handling", req.Method, r)
		}
	}()

	if !h.authed {
		return h.handleAuth(ctx, conn, req)
	}

	switch req.Method {
	case "getText":
		return h.handleGetFile(ctx, conn, req)
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}

func (h *websocketHandler) run(conn *jsonrpc2.Conn) {
	var callback = func(value interface{}) {
		if !h.authed {
			return
		}
		switch t := value.(type) {
		case state.OpenFileEvent:
			conn.Notify(context.Background(), "openFile", value)
		case state.CloseFileEvent:
			conn.Notify(context.Background(), "closeFile", value)
		case state.ReplaceTextEvent:
			conn.Notify(context.Background(), "textReplaced", value)
		case state.UpdateTextEvent:
			conn.Notify(context.Background(), "updateText", value)
		case state.ChangeViewEvent:
			conn.Notify(context.Background(), "updateView", value)
		default:
			h.logger.Println("Received unknown type from state", t)
		}
	}
	h.state.Subscribe(callback)
	defer h.state.Unsubscribe(callback)

	<-conn.DisconnectNotify()
}
