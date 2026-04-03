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

func TestSubmissionServiceCreateSubmissionPublishesJob(t *testing.T) {
	submissions := outmocks.NewMockSubmissionRepository(t)
	publisher := outmocks.NewMockEvaluationPublisher(t)
	service := NewSubmissionService(submissions, publisher)

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

func TestSubmissionServiceCreateSubmissionStillSucceedsWhenPublisherFails(t *testing.T) {
	submissions := outmocks.NewMockSubmissionRepository(t)
	publisher := outmocks.NewMockEvaluationPublisher(t)
	service := NewSubmissionService(submissions, publisher)

	userID := uuid.New()
	problemID := uuid.New()

	submissions.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*domain.Submission")).
		Run(func(_ context.Context, sub *domain.Submission) {
			sub.ID = uuid.New()
			sub.Status = domain.StatusPending
			sub.SubmittedAt = time.Now()
		}).
		Return(nil).
		Once()

	publisher.EXPECT().PublishEvaluation(mock.Anything).Return(errors.New("queue down")).Once()

	sub, err := service.CreateSubmission(context.Background(), userID, problemID, "python", "print(1)", nil)
	require.NoError(t, err)
	require.NotNil(t, sub)
	assert.Equal(t, userID, sub.UserID)
	assert.Equal(t, problemID, sub.ProblemID)
	assert.Equal(t, domain.StatusPending, sub.Status)
}

func TestSubmissionServiceGetSubmission(t *testing.T) {
	submissions := outmocks.NewMockSubmissionRepository(t)
	service := NewSubmissionService(submissions, nil)

	ctx := context.Background()
	submissionID := uuid.New()
	sub := &domain.Submission{ID: submissionID}
	submissions.EXPECT().GetByID(ctx, submissionID).Return(sub, nil)

	got, err := service.GetSubmission(ctx, submissionID)
	require.NoError(t, err)
	require.Equal(t, sub, got)
}

func TestSubmissionServiceListSubmissions(t *testing.T) {
	submissions := outmocks.NewMockSubmissionRepository(t)
	service := NewSubmissionService(submissions, nil)

	ctx := context.Background()
	userID := uuid.New()
	problemID := uuid.New()
	problemPtr := &problemID
	subList := []domain.Submission{{ID: uuid.New()}}
	submissions.EXPECT().ListByUser(ctx, userID, problemPtr, 20, 5).Return(subList, nil)

	got, err := service.ListSubmissions(ctx, userID, problemPtr, 20, 5)
	require.NoError(t, err)
	require.Equal(t, subList, got)
}

func TestSubmissionServiceCreateSubmissionReturnsCreateError(t *testing.T) {
	submissions := outmocks.NewMockSubmissionRepository(t)
	service := NewSubmissionService(submissions, nil)

	ctx := context.Background()
	userID := uuid.New()
	submissions.EXPECT().Create(ctx, mock.AnythingOfType("*domain.Submission")).Return(errors.New("create"))

	_, err := service.CreateSubmission(ctx, userID, uuid.New(), "go", "code", nil)
	require.Error(t, err)
}
