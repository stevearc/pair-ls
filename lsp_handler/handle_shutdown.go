package lsp_handler

import (
	"context"

	"github.com/sourcegraph/jsonrpc2"
)

func (h *lspHandler) handleShutdown(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	return nil, conn.Close()
}
