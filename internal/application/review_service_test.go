package application

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
	outmocks "github.com/tuannm99/judge-loop/internal/port/out/mocks"
)

func TestReviewServiceGetReviewsToday(t *testing.T) {
	reviews := outmocks.NewMockReviewRepository(t)
	service := NewReviewService(reviews)

	ctx := context.Background()
	userID := uuid.New()
	due := []outport.DueReview{{ProblemID: uuid.New()}}
	reviews.EXPECT().GetDue(ctx, userID).Return(due, nil)

	got, err := service.GetReviewsToday(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, due, got)
}
