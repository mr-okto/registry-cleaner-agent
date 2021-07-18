package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	"log"
	"registry-cleaner-agent/internal/app/agent"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath,
		"config",
		"config/agent.toml",
		"path to config file")
}

func main() {
	flag.Parse()
	config := agent.Config{}
	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Launching agent at %s", config.BindAddr)
	if err := agent.New(&config).Start(); err != nil {
		log.Fatal(err)
	}
}
