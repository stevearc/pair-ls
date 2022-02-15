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

type lspCommand struct {
	config *PairConfig
}

func NewLSPCmd(conf *PairConfig) command.Cmd {
	return &lspCommand{
		config: conf,
	}
}

func (cmd *lspCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.config.WebHostname, "hostname", cmd.config.WebHostname, "Hostname for webserver to bind to")
	fs.IntVar(&cmd.config.WebPort, "port", cmd.config.WebPort, "Port for webserver to listen on")
	fs.StringVar(&cmd.config.ForwardHost, "forward", cmd.config.ForwardHost, "Forward to relay server (use full ws://host url format)")
	return fs
}

func (cmd *lspCommand) Run(args []string) {
	f, err := os.OpenFile(cmd.config.LogFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	state := state.NewState(log.New(f, "[State]", log.Ldate|log.Ltime|log.Lshortfile))

	if cmd.config.WebPort > 0 {
		conf := server.ServerConfig{
			Password: cmd.config.WebPassword,
			KeyFile:  cmd.config.WebKeyFile,
			CertFile: cmd.config.WebCertFile,
		}
		server := server.NewServer(state, log.New(f, "[Webserver]", log.Ldate|log.Ltime|log.Lshortfile), conf)
		go server.Serve(cmd.config.WebHostname, cmd.config.WebPort)
	}

	conf := server.ServerConfig{
		CertFile: cmd.config.RelayCertFile,
	}
	lspLogger := log.New(f, "[LSP server]", log.Ldate|log.Ltime|log.Lshortfile)
	handler := lsp_handler.NewHandler(state, lspLogger, &conf, cmd.config.ForwardHost)

	lsp_handler.ListenOnStdin(handler, lspLogger, cmd.config.LogLevel)
}
