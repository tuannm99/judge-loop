package mocks

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMockEvaluationService(t *testing.T) {
	ctx := context.Background()
	subID := uuid.New()
	userID := uuid.New()

	m := NewMockEvaluationService(t)
	require.NotNil(t, m.EXPECT())
	m.EXPECT().EvaluateSubmission(mock.Anything, subID, userID, 2).Run(func(context.Context, uuid.UUID, uuid.UUID, int) {}).Return(nil)
	require.NoError(t, m.EvaluateSubmission(ctx, subID, userID, 2))
	m.EXPECT().EvaluateSubmission(mock.Anything, subID, userID, 3).RunAndReturn(func(context.Context, uuid.UUID, uuid.UUID, int) error { return nil })
	require.NoError(t, m.EvaluateSubmission(ctx, subID, userID, 3))
}
