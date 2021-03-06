package lsp_handler

import (
	"context"
	"encoding/json"
	"pair-ls/state"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type CursorMoveRequest struct {
	TextDocument lsp.TextDocumentIdentifier `json:"textDocument"`
	Cursors      []state.CursorPosition     `json:"cursors"`
}

func (h *LspHandler) handleCursorMove(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params CursorMoveRequest
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}
	filename, err := h.filenameFromURI(params.TextDocument.URI)
	if err != nil {
		return nil, nil
	}
	h.state.CursorMove(filename, params.Cursors)
	return nil, nil
}
