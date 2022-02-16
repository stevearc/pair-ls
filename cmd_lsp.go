package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"pair-ls/lsp_handler"
	"pair-ls/server"
	"pair-ls/state"
	"pair-ls/util"

	"github.com/rakyll/command"
)

type lspCommand struct {
	config       *PairConfig
	host         string
	port         int
	forwardHost  string
	signalServer string
}

func NewLSPCmd(conf *PairConfig) command.Cmd {
	return &lspCommand{
		config: conf,
	}
}

func (cmd *lspCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.host, "hostname", "", "Hostname for webserver to bind to")
	fs.IntVar(&cmd.port, "port", -1, "Port for webserver to listen on")
	addServerFlags(cmd.config, fs)
	fs.StringVar(&cmd.forwardHost, "forward", "", "Forward to relay server (use full ws:// or wss:// url format)")
	fs.StringVar(&cmd.signalServer, "signal", "", "Connect to signal server (use full ws:// or wss:// url format)")
	fs.StringVar(&cmd.config.CallToken, "call-token", cmd.config.CallToken, "WebRTC token copied from static server")
	fs.StringVar(&cmd.config.Client.CertFile, "client-cert", cmd.config.Client.CertFile, "Client certificate used to connect to relay/signal server")
	fs.StringVar(&cmd.config.Client.KeyFile, "client-key", cmd.config.Client.KeyFile, "Client key used to connect to relay/signal server")
	return fs
}

func (cmd *lspCommand) Run(args []string) {
	f, err := os.OpenFile(cmd.config.LogFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	state := state.NewState(log.New(f, "[State]", log.Ldate|log.Ltime|log.Lshortfile))

	if cmd.port > 0 {
		server := server.NewServer(state, log.New(f, "[Webserver]", log.Ldate|log.Ltime|log.Lshortfile), cmd.config.Server)
		go server.Serve(cmd.host, cmd.port)
	}

	conf := lsp_handler.HandlerConfig{
		RelayServer:   cmd.forwardHost,
		SignalServer:  cmd.signalServer,
		StaticRTCSite: cmd.config.StaticRTCSite,
		ClientAuth:    cmd.config.Client,
	}
	lspLogger := log.New(f, "[LSP server]", log.Ldate|log.Ltime|log.Lshortfile)
	handler := lsp_handler.NewHandler(state, lspLogger, &conf)

	if cmd.port > 0 {
		hostname := cmd.host
		if hostname == "" {
			hostname = "localhost"
		}
		handler.SendShareString(util.CreateShareURL(fmt.Sprintf("%s:%d", hostname, cmd.port), ""))
	}
	handler.ListenOnStdin(lspLogger, cmd.config.LogLevel, cmd.config.CallToken)
}
