package apiserver

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	postgres "github.com/tuannm99/judge-loop/internal/infrastructure/postgres"
	inmocks "github.com/tuannm99/judge-loop/internal/port/in/mocks"
)

func TestNew(t *testing.T) {
	api := New(&postgres.DB{}, uuid.New(), nil)
	require.NotNil(t, api)
	require.NotNil(t, api.Problems)
	require.NotNil(t, api.Submissions)
	require.NotNil(t, api.Progress)
	require.NotNil(t, api.Timers)
	require.NotNil(t, api.Reviews)
	require.NotNil(t, api.Registry)
}

func TestNewWithService(t *testing.T) {
	service := inmocks.NewMockAPIService(t)
	api := NewWithService(service, uuid.New())
	require.NotNil(t, api)
}
