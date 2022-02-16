package lsp_handler

import (
	"context"
	"encoding/json"
	"pair-ls/util"
	"path/filepath"
	"reflect"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func (h *LspHandler) handleInitialize(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.InitializeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	rootPath, err := util.FromURI(params.RootURI)
	if err != nil {
		return nil, err
	}
	h.rootPath = filepath.Clean(rootPath)

	exp := reflect.ValueOf(params.Capabilities.Experimental)
	if exp.IsValid() && exp.Kind() == reflect.Map {
		cursor_capabilities := exp.MapIndex(reflect.ValueOf("cursor"))
		if cursor_capabilities.Kind() == reflect.Interface {
			cursor := reflect.ValueOf(cursor_capabilities.Interface())
			if cursor.IsValid() && cursor.Kind() == reflect.Map {
				key := reflect.ValueOf("position")
				support_val := cursor.MapIndex(key)
				if support_val.IsValid() {
					support := reflect.ValueOf(support_val.Interface())
					if support.IsValid() && support.Kind() == reflect.Bool {
						h.clientSendsCursor = support.Bool()
					}
				}
			}
		}
	}

	return lsp.InitializeResult{
		Capabilities: lsp.ServerCapabilities{
			HoverProvider: !h.clientSendsCursor,
			TextDocumentSync: &lsp.TextDocumentSyncOptionsOrKind{
				Options: &lsp.TextDocumentSyncOptions{
					OpenClose:         true,
					Change:            lsp.TDSKIncremental,
					WillSave:          false,
					WillSaveWaitUntil: false,
				},
			},
		},
	}, nil
}
