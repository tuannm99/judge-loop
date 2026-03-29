// local-agent is the lightweight HTTP daemon running on the developer's machine.
// It bridges the Neovim plugin and the remote api-server.
// Listens on 127.0.0.1:7070 (localhost only).
package main

import (
	"log"
)

func main() {
	cfg := LoadConfig()

	log.Printf("local-agent starting — server: %s, port: %d", cfg.ServerURL, cfg.Port)

	srv := NewServer(cfg)
	if err := srv.Run(); err != nil {
		log.Fatalf("local-agent: %v", err)
	}
}
