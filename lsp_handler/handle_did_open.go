package lsp_handler

import (
	"context"
	"encoding/json"
	"pair-ls/util"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func (h *lspHandler) handleTextDocumentDidOpen(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.DidOpenTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}
	filename, err := util.FromURI(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	h.state.OpenFile(filename, params.TextDocument.Text, params.TextDocument.LanguageID, !h.clientSendsCursor)
	return nil, nil
}
