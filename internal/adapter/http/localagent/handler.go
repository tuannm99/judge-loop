package localagent

import (
	localtimer "github.com/tuannm99/judge-loop/internal/infrastructure/localtimer"
)

// Handler holds the dependencies shared by all request handlers.
// All handler methods live in handler_*.go files in the same package.
type Handler struct {
	client       *APIClient
	timer        *localtimer.LocalTimer
	userID       string
	registryPath string
}

// NewHandler creates a Handler wired to the given client and config.
func NewHandler(client *APIClient, userID, registryPath string) *Handler {
	return &Handler{
		client:       client,
		timer:        localtimer.New(),
		userID:       userID,
		registryPath: registryPath,
	}
}
