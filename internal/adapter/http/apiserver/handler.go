// Package apiserver contains Gin HTTP handlers for the api-server.
package apiserver

import (
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	queueadapter "github.com/tuannm99/judge-loop/internal/adapter/queue"
	storageadapter "github.com/tuannm99/judge-loop/internal/adapter/storage"
	application "github.com/tuannm99/judge-loop/internal/application"
	postgres "github.com/tuannm99/judge-loop/internal/infrastructure/postgres"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
)

// Handler holds the data stores and per-request dependencies for all handlers.
// UserID is the single dev user — no authentication in v1.
type Handler struct {
	Service inport.APIService
	UserID  uuid.UUID
}

// New creates a Handler wired to the given database and queue client.
func New(db *postgres.DB, userID uuid.UUID, queueClient *asynq.Client) *Handler {
	return &Handler{
		Service: application.NewAPIService(
			storageadapter.NewProblemRepository(db),
			storageadapter.NewSubmissionRepository(db),
			storageadapter.NewSessionRepository(db),
			storageadapter.NewReviewRepository(db),
			storageadapter.NewRegistryRepository(db),
			queueadapter.NewEvaluationPublisher(queueClient),
		),
		UserID: userID,
	}
}
