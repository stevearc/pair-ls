package lsp_handler

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"pair-ls/auth"
	"pair-ls/util"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sourcegraph/jsonrpc2"
)

func (h *LspHandler) forward(config ClientAuthConfig) {
	forwardHost := h.config.RelayServer
	if !strings.HasSuffix(forwardHost, "/relay") {
		forwardHost = forwardHost + "/relay"
	}
	h.logger.Println("Connecting to relay server", forwardHost)
	u, err := url.Parse(forwardHost)
	if err != nil {
		h.logger.Fatal("Invalid relay server:", err)
	}
	c, err := wsDialServer(u.String(), config)
	if err != nil {
		h.logger.Fatal("Websocket dial error:", err)
	}

	conn := jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(util.WrapWebsocket(c), jsonrpc2.PlainObjectCodec{}),
		jsonrpc2.HandlerWithError(func(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
			return nil, nil
		}),
	)
	h.SendShareString(util.CreateShareURL(h.config.RelayServer, ""))
	func() {
		defer c.Close()
		for {
			req := <-h.forwardChan
			conn.Notify(context.Background(), req.Method, req.Params)
		}
	}()
}

func wsDialServer(urlStr string, config ClientAuthConfig) (*websocket.Conn, error) {
	var tlsConfig *tls.Config = nil
	if config.CertFile != "" {
		var err error
		tlsConfig, err = auth.LoadTLSConfig(config.CertFile, config.KeyFile)
		if err != nil {
			return nil, err
		}
	}
	header := http.Header{}
	if config.Password != "" {
		header.Add("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(config.Password))))
	}
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		TLSClientConfig:  tlsConfig,
	}

	c, _, err := dialer.Dial(urlStr, header)
	return c, err
}
