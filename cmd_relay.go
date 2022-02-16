package main

import (
	"flag"
	"log"
	"os"
	"pair-ls/lsp_handler"
	"pair-ls/server"
	"pair-ls/state"

	"github.com/rakyll/command"
)

type relayCommand struct {
	config *PairConfig
	host   string
	port   int
}

func NewRelayCmd(conf *PairConfig) command.Cmd {
	return &relayCommand{
		config: conf,
	}
}

func (cmd *relayCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.host, "host", "", "Hostname to bind to")
	fs.IntVar(&cmd.port, "port", -1, "Port to listen on")
	addServerFlags(cmd.config, fs)
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

	lspLogger := log.New(f, "[LSP server]", log.Ldate|log.Ltime|log.Lshortfile)
	handler := lsp_handler.NewHandler(state, lspLogger, &lsp_handler.HandlerConfig{})

	srv := server.NewServer(state, log.New(f, "[Relay]", log.Ldate|log.Ltime|log.Lshortfile), cmd.config.Server)

	relayConf := server.RelayConfig{
		Persist: cmd.config.RelayPersist,
	}
	srv.AddRelayServer(handler.GetRPCHandler(), relayConf)
	srv.Serve(cmd.host, cmd.port)
}
