package apiserver

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	inmocks "github.com/tuannm99/judge-loop/internal/port/in/mocks"
)

func TestNew(t *testing.T) {
	api := New(
		inmocks.NewMockProblemService(t),
		inmocks.NewMockSubmissionService(t),
		inmocks.NewMockProgressService(t),
		inmocks.NewMockTimerService(t),
		inmocks.NewMockReviewService(t),
		inmocks.NewMockRegistryService(t),
		uuid.New(),
	)
	require.NotNil(t, api)
	require.NotNil(t, api.Problems)
	require.NotNil(t, api.Submissions)
	require.NotNil(t, api.Progress)
	require.NotNil(t, api.Timers)
	require.NotNil(t, api.Reviews)
	require.NotNil(t, api.Registry)
}
