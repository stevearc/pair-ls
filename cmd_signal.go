package main

import (
	"flag"
	"log"
	"os"
	"pair-ls/server"

	"github.com/rakyll/command"
)

type signalCommand struct {
	config *PairConfig
	host   string
	port   int
}

func NewSignalCmd(conf *PairConfig) command.Cmd {
	return &signalCommand{
		config: conf,
	}
}

func (cmd *signalCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.host, "host", "", "Hostname for webserver to bind to")
	fs.IntVar(&cmd.port, "port", -1, "Port for webserver to listen on")
	addServerFlags(cmd.config, fs)
	return fs
}

func (cmd *signalCommand) Run(args []string) {
	f, err := os.OpenFile(cmd.config.LogFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	logger := log.New(f, "[Signal]", log.Ldate|log.Ltime|log.Lshortfile)
	srv := server.NewServer(nil, logger, cmd.config.Server)
	srv.MakeSignalServer()
	srv.Serve(cmd.host, cmd.port)
}
