// local-agent is the lightweight HTTP daemon running on the developer's machine.
// It bridges the Neovim plugin and the remote api-server.
// Listens on 127.0.0.1:7070 (localhost only).
package main

import (
	"log"

	"github.com/tuannm99/judge-loop/internal/config"
	"github.com/tuannm99/judge-loop/internal/di"
)

func main() {
	cfg, err := config.LoadLocalAgent()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	log.Printf("local-agent starting — server: %s, port: %d", cfg.ServerURL, cfg.Port)

	di.NewLocalAgent(cfg).Run()
}
