package main

import (
	"flag"
	"log"
	"os"
	"pair-ls/lsp_handler"
	"pair-ls/relay_server"
	"pair-ls/server"
	"pair-ls/state"

	"github.com/rakyll/command"
)

type relayCommand struct {
	config *PairConfig
}

func NewRelayCmd(conf *PairConfig) command.Cmd {
	return &relayCommand{
		config: conf,
	}
}

func (cmd *relayCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.config.WebHostname, "web-hostname", cmd.config.WebHostname, "Hostname for webserver to bind to")
	fs.IntVar(&cmd.config.WebPort, "web-port", cmd.config.WebPort, "Port for webserver to listen on")
	fs.StringVar(&cmd.config.RelayHostname, "relay-hostname", cmd.config.RelayHostname, "Hostname for relay server to bind to")
	fs.IntVar(&cmd.config.RelayPort, "relay-port", cmd.config.RelayPort, "Port for relay server to listen on")
	fs.BoolVar(&cmd.config.RelayPersist, "persist", cmd.config.RelayPersist, "Keep file data even after all forwarding servers disconnect")
	return fs
}

func (cmd *relayCommand) Run(args []string) {
	f, err := os.OpenFile(cmd.config.LogFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	state := state.NewState(log.New(f, "[State]", log.Ldate|log.Ltime|log.Lshortfile))

	conf := server.ServerConfig{
		Password: cmd.config.WebPassword,
		KeyFile:  cmd.config.KeyFile,
		CertFile: cmd.config.CertFile,
	}
	if cmd.config.WebPort > 0 {
		server := server.NewServer(state, log.New(f, "[Webserver]", log.Ldate|log.Ltime|log.Lshortfile), conf)
		go server.Serve(cmd.config.WebHostname, cmd.config.WebPort)
	}

	lspLogger := log.New(f, "[LSP server]", log.Ldate|log.Ltime|log.Lshortfile)
	handler := lsp_handler.NewHandler(state, lspLogger, &conf, "")

	relayConf := relay_server.ServerConfig{
		KeyFile:  cmd.config.KeyFile,
		CertFile: cmd.config.CertFile,
		Persist:  cmd.config.RelayPersist,
	}
	relayServer := relay_server.NewServer(handler, log.New(f, "[Relay]", log.Ldate|log.Ltime|log.Lshortfile), state, relayConf)
	relayServer.Serve(cmd.config.RelayHostname, cmd.config.RelayPort)
}
