package lsp_handler

import (
	"context"
	"log"
	"net/url"
	"pair-ls/auth"
	"pair-ls/util"
	"strings"

	"github.com/sourcegraph/jsonrpc2"
)

func forward(forwardHost string, logger *log.Logger, forwardChan chan *jsonrpc2.Request, certFile string, keyFile string) {
	if certFile == "" || keyFile == "" {
		msg := "Forwarding to relay server requires a keyFile and certFile"
		logger.Println(msg)
		log.Fatalln(msg)
	}
	if !strings.HasPrefix(forwardHost, "wss://") {
		forwardHost = "wss://" + forwardHost
	}
	logger.Println("Connecting to relay server", forwardHost)
	u, err := url.Parse(forwardHost)
	if err != nil {
		log.Fatal("Invalid forwardHost:", err)
	}
	dialer, err := auth.GetTLSDialer(certFile, keyFile)
	if err != nil {
		log.Fatal("Cert error:", err)
	}
	c, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Websocket dial error:", err)
	}

	conn := jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(util.WrapWebsocket(c), jsonrpc2.PlainObjectCodec{}),
		jsonrpc2.HandlerWithError(func(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
			return nil, nil
		}),
	)
	func() {
		defer c.Close()
		for {
			req := <-forwardChan
			conn.Notify(context.Background(), req.Method, req.Params)
		}
	}()
}
