// Package handler contains all Gin HTTP handler methods for the api-server.
package handler

import (
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/tuannm99/judge-loop/internal/storage"
)

// Handler holds the data stores and per-request dependencies for all handlers.
// UserID is the single dev user — no authentication in v1.
type Handler struct {
	Problems    *storage.ProblemStore
	Submissions *storage.SubmissionStore
	Sessions    *storage.SessionStore
	Reviews     *storage.ReviewStore
	Registry    *storage.RegistryStore
	UserID      uuid.UUID
	Queue       *asynq.Client
}

// New creates a Handler wired to the given database and queue client.
func New(db *storage.DB, userID uuid.UUID, queueClient *asynq.Client) *Handler {
	return &Handler{
		Problems:    storage.NewProblemStore(db),
		Submissions: storage.NewSubmissionStore(db),
		Sessions:    storage.NewSessionStore(db),
		Reviews:     storage.NewReviewStore(db),
		Registry:    storage.NewRegistryStore(db),
		UserID:      userID,
		Queue:       queueClient,
	}
}
