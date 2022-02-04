package lsp_handler

import (
	"context"
	"encoding/json"
	"pair-ls/util"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func (h *lspHandler) handleTextDocumentDidChange(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.DidChangeTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}
	filename, err := util.FromURI(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}

	for _, change := range params.ContentChanges {
		if &change.Range == nil {
			// TODO
			h.logger.Println("We don't support incremental changes yet")
		} else {
			h.changeTextChan <- TextChange{
				Filename: filename,
				Text:     change.Text,
			}
		}
	}

	return nil, nil
}
