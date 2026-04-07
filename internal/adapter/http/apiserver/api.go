// Package apiserver contains Gin HTTP handlers for the api-server.
package apiserver

import (
	"github.com/google/uuid"
	inport "github.com/tuannm99/judge-loop/internal/port/in"
)

type ProblemsAPI struct {
	userID  uuid.UUID
	service inport.ProblemService
}

type SubmissionsAPI struct {
	userID  uuid.UUID
	service inport.SubmissionService
}

type ProgressAPI struct {
	userID  uuid.UUID
	service inport.ProgressService
}

type TimersAPI struct {
	userID  uuid.UUID
	service inport.TimerService
}

type ReviewsAPI struct {
	userID  uuid.UUID
	service inport.ReviewService
}

type RegistryAPI struct {
	service inport.RegistryService
}

type MissionsAPI struct {
	userID  uuid.UUID
	service inport.MissionService
}

// API groups the feature-specific HTTP handlers for route registration.
type API struct {
	Problems    *ProblemsAPI
	Submissions *SubmissionsAPI
	Progress    *ProgressAPI
	Timers      *TimersAPI
	Reviews     *ReviewsAPI
	Registry    *RegistryAPI
	Missions    *MissionsAPI
}

// New creates an API from the given capability services.
func New(
	problems inport.ProblemService,
	submissions inport.SubmissionService,
	progress inport.ProgressService,
	timers inport.TimerService,
	reviews inport.ReviewService,
	registry inport.RegistryService,
	missions inport.MissionService,
	userID uuid.UUID,
) *API {
	return &API{
		Problems:    &ProblemsAPI{userID: userID, service: problems},
		Submissions: &SubmissionsAPI{userID: userID, service: submissions},
		Progress:    &ProgressAPI{userID: userID, service: progress},
		Timers:      &TimersAPI{userID: userID, service: timers},
		Reviews:     &ReviewsAPI{userID: userID, service: reviews},
		Registry:    &RegistryAPI{service: registry},
		Missions:    &MissionsAPI{userID: userID, service: missions},
	}
}
