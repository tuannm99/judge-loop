package application

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
	outmocks "github.com/tuannm99/judge-loop/internal/port/out/mocks"
)

func TestTimerServiceDelegates(t *testing.T) {
	sessions := outmocks.NewMockSessionRepository(t)
	service := NewTimerService(sessions)

	ctx := context.Background()
	userID := uuid.New()
	problemID := uuid.New()
	problemPtr := &problemID
	timer := &domain.TimerSession{ID: uuid.New()}

	sessions.EXPECT().StartTimer(ctx, userID, problemPtr).Return(timer, nil)
	gotTimer, err := service.StartTimer(ctx, userID, problemPtr)
	require.NoError(t, err)
	require.Equal(t, timer, gotTimer)

	sessions.EXPECT().StopTimer(ctx, userID).Return(timer, nil)
	gotTimer, err = service.StopTimer(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, timer, gotTimer)

	sessions.EXPECT().ActiveTimer(ctx, userID).Return(timer, nil)
	gotTimer, err = service.CurrentTimer(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, timer, gotTimer)
}
