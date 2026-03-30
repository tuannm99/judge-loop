package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
	outmocks "github.com/tuannm99/judge-loop/internal/port/out/mocks"
)

func TestAPIServiceCreateSubmissionPublishesJob(t *testing.T) {
	problems := outmocks.NewMockProblemRepository(t)
	submissions := outmocks.NewMockSubmissionRepository(t)
	sessions := outmocks.NewMockSessionRepository(t)
	reviews := outmocks.NewMockReviewRepository(t)
	registry := outmocks.NewMockRegistryRepository(t)
	publisher := outmocks.NewMockEvaluationPublisher(t)

	service := NewAPIService(problems, submissions, sessions, reviews, registry, publisher)

	userID := uuid.New()
	problemID := uuid.New()

	submissions.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(sub *domain.Submission) bool {
			sub.ID = uuid.New()
			sub.Status = domain.StatusPending
			sub.SubmittedAt = time.Now()
			return sub.UserID == userID && sub.ProblemID == problemID && sub.Language == domain.LanguagePython
		})).
		Return(nil).
		Once()

	publisher.EXPECT().
		PublishEvaluation(mock.MatchedBy(func(job outport.EvaluateSubmissionJob) bool {
			return job.UserID == userID.String() && job.SubmissionID != ""
		})).
		Return(nil).
		Once()

	sub, err := service.CreateSubmission(context.Background(), userID, problemID, "python", "print(1)", nil)
	require.NoError(t, err)
	require.NotNil(t, sub)
	assert.Equal(t, userID, sub.UserID)
	assert.Equal(t, problemID, sub.ProblemID)
	assert.Equal(t, domain.LanguagePython, sub.Language)
}

func TestAPIServiceCreateSubmissionReturnsPublisherError(t *testing.T) {
	problems := outmocks.NewMockProblemRepository(t)
	submissions := outmocks.NewMockSubmissionRepository(t)
	sessions := outmocks.NewMockSessionRepository(t)
	reviews := outmocks.NewMockReviewRepository(t)
	registry := outmocks.NewMockRegistryRepository(t)
	publisher := outmocks.NewMockEvaluationPublisher(t)

	service := NewAPIService(problems, submissions, sessions, reviews, registry, publisher)

	submissions.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*domain.Submission")).
		Run(func(_ context.Context, sub *domain.Submission) {
			sub.ID = uuid.New()
			sub.Status = domain.StatusPending
			sub.SubmittedAt = time.Now()
		}).
		Return(nil).
		Once()

	publisher.EXPECT().
		PublishEvaluation(mock.Anything).
		Return(errors.New("queue down")).
		Once()

	_, err := service.CreateSubmission(context.Background(), uuid.New(), uuid.New(), "python", "print(1)", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "queue down")
}

func TestAPIServiceGetProblemFallsBackToSlug(t *testing.T) {
	problems := outmocks.NewMockProblemRepository(t)
	submissions := outmocks.NewMockSubmissionRepository(t)
	sessions := outmocks.NewMockSessionRepository(t)
	reviews := outmocks.NewMockReviewRepository(t)
	registry := outmocks.NewMockRegistryRepository(t)
	publisher := outmocks.NewMockEvaluationPublisher(t)

	service := NewAPIService(problems, submissions, sessions, reviews, registry, publisher)

	want := &domain.Problem{Slug: "two-sum"}

	problems.EXPECT().
		GetBySlug(mock.Anything, "two-sum").
		Return(want, nil).
		Once()

	got, err := service.GetProblem(context.Background(), "two-sum")
	require.NoError(t, err)
	assert.Same(t, want, got)
}
