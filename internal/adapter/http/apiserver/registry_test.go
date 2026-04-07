package apiserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
	inmocks "github.com/tuannm99/judge-loop/internal/port/in/mocks"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

func TestRegistryHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := inmocks.NewMockRegistryService(t)
	api := New(nil, nil, nil, nil, nil, service, nil, uuid.New())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/registry/sync", bytes.NewBufferString("{"))
	c.Request.Header.Set("Content-Type", "application/json")
	api.Registry.SyncRegistry(c)
	require.Equal(t, http.StatusBadRequest, w.Code)

	updatedAt := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	body, err := json.Marshal(registrySyncRequest{
		Version:   "v1",
		UpdatedAt: updatedAt,
		Problems:  []domain.ProblemManifest{{Slug: "two-sum"}},
		Manifests: []domain.ManifestRef{{Name: "main", Path: "providers/main.json"}},
	})
	require.NoError(t, err)

	service.EXPECT().SyncRegistry(mock.Anything, "v1", updatedAt, mock.Anything, mock.Anything).Return(1, nil).Once()
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/registry/sync", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	api.Registry.SyncRegistry(c)
	require.Equal(t, http.StatusOK, w.Code)

	service.EXPECT().
		SyncRegistry(mock.Anything, "v1", updatedAt, mock.Anything, mock.Anything).
		Return(0, errors.New("boom")).
		Once()
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/registry/sync", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	api.Registry.SyncRegistry(c)
	require.Equal(t, http.StatusInternalServerError, w.Code)

	service.EXPECT().GetRegistryVersion(mock.Anything).Return(nil, nil).Once()
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/registry/version", nil)
	api.Registry.GetRegistryVersion(c)
	require.Equal(t, http.StatusOK, w.Code)

	service.EXPECT().
		GetRegistryVersion(mock.Anything).
		Return(&outport.RegistryVersion{Version: "v1", SyncedAt: time.Now()}, nil).
		Once()
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/registry/version", nil)
	api.Registry.GetRegistryVersion(c)
	require.Equal(t, http.StatusOK, w.Code)

	service.EXPECT().GetRegistryVersion(mock.Anything).Return(nil, errors.New("boom")).Once()
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/registry/version", nil)
	api.Registry.GetRegistryVersion(c)
	require.Equal(t, http.StatusInternalServerError, w.Code)
}
