package main

import (
	"io/ioutil"
	"log"
	"pair-ls/lsp_handler"
	"pair-ls/server"

	"github.com/BurntSushi/toml"
	"github.com/rakyll/command"

	"flag"
	"os"
	"path/filepath"
)

func main() {
	config_home := os.Getenv("XDG_CONFIG_HOME")
	if config_home == "" {
		config_home = filepath.Join(os.Getenv("HOME"), ".config")
	}

	confFile := filepath.Join(config_home, "pair-ls.toml")
	flag.String("config", confFile, "Path to config file")

	for i, flag := range os.Args {
		if flag == "-config" && i+1 < len(os.Args) {
			confFile = os.Args[i+1]
			break
		}
	}

	config, err := readConfig(confFile)
	if err != nil {
		log.Panicln("Could not read config file", err)
	}

	flag.StringVar(&config.LogFile, "logfile", config.LogFile, "Logs will be written here")
	flag.IntVar(&config.LogLevel, "loglevel", config.LogLevel, "Log detail")
	command.On("lsp", "Run the LSP server", NewLSPCmd(config), []string{})
	command.On("relay", "Run a relay server", NewRelayCmd(config), []string{"port"})
	command.On("signal", "Run a signal server for making WebRTC connections", NewSignalCmd(config), []string{"port"})
	command.On("cert", "Generate certificates for relay server", NewCertCmd(config), []string{})
	command.ParseAndRun()
}

func addServerFlags(config *PairConfig, fs *flag.FlagSet) {
	fs.StringVar(&config.Server.KeyFile, "key", config.Server.KeyFile, "Path to the TLS key file for the webserver")
	fs.StringVar(&config.Server.CertFile, "cert", config.Server.CertFile, "Path to the TLS certificate file for the webserver")
	fs.StringVar(&config.Server.ClientCAs, "client-ca", config.Server.ClientCAs, "Path to certificate pool used to auth clients with -require-client-cert")
	fs.BoolVar(&config.Server.RequireClientCert, "require-client-cert", config.Server.RequireClientCert, "Require pair-ls LSP clients to auth with a client certificate")
}

func readConfig(filename string) (*PairConfig, error) {
	cache_home := os.Getenv("XDG_CACHE_HOME")
	if cache_home == "" {
		cache_home = filepath.Join(os.Getenv("HOME"), ".cache")
	}
	config := PairConfig{
		LogFile:       filepath.Join(cache_home, "pair-ls.log"),
		LogLevel:      1,
		StaticRTCSite: "https://code.stevearc.com/",
	}
	content, err := ioutil.ReadFile(filename)
	if err == nil {
		_, err := toml.Decode(string(content), &config)
		if err != nil {
			return nil, err
		}
	}
	webPass := os.Getenv("PAIR_WEB_PASS")
	if webPass != "" {
		config.Server.WebPassword = webPass
	}

	return &config, nil
}

type PairConfig struct {
	LogFile       string                       `json:"logFile"`
	LogLevel      int                          `json:"logLevel"`
	Server        server.WebServerConfig       `json:"server"`
	Client        lsp_handler.ClientAuthConfig `json:"client"`
	RelayPersist  bool                         `json:"relayPersist"`
	CallToken     string                       `json:"callToken"`
	StaticRTCSite string                       `json:"staticRTCSite"`
}
