package main

import (
	"log"
	"os"
)

func main() {
	addr := os.Getenv("SMTP_LISTEN_ADDR")
	if addr == "" {
		addr = ":2525"
	}

	log.Printf("ComplyMail SMTP Proxy listening on %s", addr)
	// TODO: start SMTP server, accept connections, parse emails,
	//       call backend API for scanning, relay upstream.
	select {} // block forever (placeholder)
}
