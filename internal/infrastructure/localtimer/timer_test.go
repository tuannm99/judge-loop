package localtimer

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestLocalTimerLifecycle(t *testing.T) {
	timer := New()
	require.Nil(t, timer.Active())
	require.Zero(t, timer.ElapsedSecs())

	problemID := uuid.New()
	entry := timer.Start(&problemID)
	require.NotEqual(t, uuid.Nil, entry.ID)
	require.NotNil(t, entry.ProblemID)

	active := timer.Active()
	require.NotNil(t, active)
	require.Equal(t, entry.ID, active.ID)

	active.StartedAt = active.StartedAt.Add(-3 * time.Second)
	timer.active = active
	require.GreaterOrEqual(t, timer.ElapsedSecs(), 3)

	stopped, ok := timer.Stop()
	require.True(t, ok)
	require.Equal(t, entry.ID, stopped.ID)
	require.Nil(t, timer.Active())

	zero, ok := timer.Stop()
	require.False(t, ok)
	require.Equal(t, Entry{}, zero)
}
