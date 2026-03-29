package main

import (
	"github.com/tuannm99/judge-loop/internal/timer"
)

// Handler holds the dependencies shared by all request handlers.
// All handler methods live in handler_*.go files in the same package.
type Handler struct {
	client *APIClient
	timer  *timer.LocalTimer
	userID string
}

// NewHandler creates a Handler wired to the given client and config.
func NewHandler(client *APIClient, cfg Config) *Handler {
	return &Handler{
		client: client,
		timer:  timer.New(),
		userID: cfg.UserID,
	}
}
