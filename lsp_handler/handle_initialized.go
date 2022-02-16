package lsp_handler

import (
	"context"

	"github.com/sourcegraph/jsonrpc2"
)

func (h *LspHandler) handleInitialized(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	h.initialized = true
	for _, notif := range h.pendingNotifs {
		err := h.lspConn.Notify(context.Background(), notif.method, notif.params)
		if err != nil {
			h.logger.Println("Error notifying LSP client")
		}
	}
	h.pendingNotifs = h.pendingNotifs[:0]
	return nil, nil
}
