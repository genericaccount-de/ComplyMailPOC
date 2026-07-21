package main

import (
	"flag"
	"log"

	"github.com/genericaccount-de/comply-mail-poc/smtp-proxy/internal/config"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to YAML configuration file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	log.Printf("ComplyMail SMTP Proxy listening on %s (upstream=%s:%d, backend=%s)",
		cfg.ListenAddr, cfg.Upstream.Host, cfg.Upstream.Port, cfg.Backend.APIURL)
	// TODO: start SMTP server, accept connections, parse emails,
	//       call backend API for scanning, relay upstream.
	select {} // block forever (placeholder)
}
