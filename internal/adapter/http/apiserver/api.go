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

type deps struct {
	userID      uuid.UUID
	problems    inport.ProblemService
	submissions inport.SubmissionService
	progress    inport.ProgressService
	timers      inport.TimerService
	reviews     inport.ReviewService
	registry    inport.RegistryService
}

type ProblemsAPI struct{ deps deps }

type SubmissionsAPI struct{ deps deps }

type ProgressAPI struct{ deps deps }

type TimersAPI struct{ deps deps }

type ReviewsAPI struct{ deps deps }

type RegistryAPI struct{ deps deps }

// API groups the feature-specific HTTP handlers for route registration.
type API struct {
	Problems    *ProblemsAPI
	Submissions *SubmissionsAPI
	Progress    *ProgressAPI
	Timers      *TimersAPI
	Reviews     *ReviewsAPI
	Registry    *RegistryAPI
}

// New creates an API wired to the given database and queue client.
func New(db *postgres.DB, userID uuid.UUID, queueClient *asynq.Client) *API {
	service := application.NewAPIService(
		storageadapter.NewProblemRepository(db),
		storageadapter.NewSubmissionRepository(db),
		storageadapter.NewSessionRepository(db),
		storageadapter.NewReviewRepository(db),
		storageadapter.NewRegistryRepository(db),
		queueadapter.NewEvaluationPublisher(queueClient),
	)
	return NewWithService(service, userID)
}

// NewWithService creates an API from a prebuilt service. Useful for tests.
func NewWithService(service inport.APIService, userID uuid.UUID) *API {
	d := deps{
		userID:      userID,
		problems:    service,
		submissions: service,
		progress:    service,
		timers:      service,
		reviews:     service,
		registry:    service,
	}
	return &API{
		Problems:    &ProblemsAPI{deps: d},
		Submissions: &SubmissionsAPI{deps: d},
		Progress:    &ProgressAPI{deps: d},
		Timers:      &TimersAPI{deps: d},
		Reviews:     &ReviewsAPI{deps: d},
		Registry:    &RegistryAPI{deps: d},
	}
}
