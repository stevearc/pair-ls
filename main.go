package main

import (
	"encoding/json"
	"io/ioutil"
	"log"

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

	confFile := filepath.Join(config_home, "pair-ls.json")
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
	flag.StringVar(&config.KeyFile, "key", config.KeyFile, "Path to the TLS key file")
	flag.StringVar(&config.CertFile, "cert", config.CertFile, "Path to the TLS cert file")
	command.On("lsp", "Run the LSP server", NewLSPCmd(config), []string{})
	command.On("relay", "Run a relay server", NewRelayCmd(config), []string{})
	command.On("cert", "Generate certificates for web and relay server", NewCertCmd(config), []string{})
	command.ParseAndRun()
}

func readConfig(filename string) (*PairConfig, error) {
	cache_home := os.Getenv("XDG_CACHE_HOME")
	if cache_home == "" {
		cache_home = filepath.Join(os.Getenv("HOME"), ".cache")
	}
	config := PairConfig{
		LogFile:  filepath.Join(cache_home, "pair-ls.log"),
		LogLevel: 1,
	}
	content, err := ioutil.ReadFile(filename)
	if err == nil {
		if err := json.Unmarshal(content, &config); err != nil {
			return nil, err
		}
	}
	webPass := os.Getenv("PAIR_WEB_PASS")
	if webPass != "" {
		config.WebPassword = webPass
	}

	return &config, nil
}

type PairConfig struct {
	LogFile       string `json:"logFile"`
	LogLevel      int    `json:"logLevel"`
	KeyFile       string `json:"keyFile"`
	CertFile      string `json:"certFile"`
	WebHostname   string `json:"webHostname"`
	WebPort       int    `json:"webPort"`
	WebPassword   string `json:"webPassword"`
	ForwardHost   string `json:"forwardHost"`
	RelayHostname string `json:"relayHostname"`
	RelayPort     int    `json:"relayPort"`
	RelayPersist  bool   `json:"relayPersist"`
}
