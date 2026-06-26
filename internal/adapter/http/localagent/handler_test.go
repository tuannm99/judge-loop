package localagent

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewHandler(t *testing.T) {
	h := NewHandler(&APIClient{}, "user-1", "/tmp/registry")
	require.NotNil(t, h.client)
	require.NotNil(t, h.timer)
	require.Equal(t, "user-1", h.userID)
}
