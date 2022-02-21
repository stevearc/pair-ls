package lsp_handler

import (
	"context"
	"encoding/json"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func (h *LspHandler) handleTextDocumentDidChange(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.DidChangeTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}
	filename, err := h.filenameFromURI(params.TextDocument.URI)
	if err != nil {
		return nil, nil
	}

	for _, change := range params.ContentChanges {
		if change.Range == nil {
			h.changeTextChan <- TextChange{
				Filename: filename,
				Text:     change.Text,
			}
			return nil, nil
		}
	}
	h.state.ReplaceTextRanges(filename, params.ContentChanges)

	return nil, nil
}
