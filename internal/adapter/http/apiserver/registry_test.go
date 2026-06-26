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

func TestRegistryAPISyncRegistry(t *testing.T) {
	gin.SetMode(gin.TestMode)

	updatedAt := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	body, err := json.Marshal(registrySyncRequest{
		Version:   "v1",
		UpdatedAt: updatedAt,
		Problems:  []domain.ProblemManifest{{Slug: "two-sum"}},
		Manifests: []domain.ManifestRef{{Name: "main", Path: "providers/main.json"}},
	})
	require.NoError(t, err)

	cases := []struct {
		name       string
		body       []byte
		synced     int
		err        error
		wantStatus int
		wantCall   bool
	}{
		{name: "bad request", body: []byte("{"), wantStatus: http.StatusBadRequest},
		{name: "success", body: body, synced: 1, wantStatus: http.StatusOK, wantCall: true},
		{
			name:       "service error",
			body:       body,
			err:        errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantCall:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := inmocks.NewMockRegistryService(t)
			api := New(nil, nil, nil, nil, nil, service, nil, uuid.New())
			if tc.wantCall {
				service.EXPECT().
					SyncRegistry(mock.Anything, "v1", updatedAt, mock.Anything, mock.Anything).
					Return(tc.synced, tc.err).
					Once()
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/registry/sync", bytes.NewReader(tc.body))
			c.Request.Header.Set("Content-Type", "application/json")
			api.Registry.SyncRegistry(c)
			require.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestRegistryAPIGetRegistryVersion(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name       string
		version    *outport.RegistryVersion
		err        error
		wantStatus int
	}{
		{name: "empty", wantStatus: http.StatusOK},
		{
			name:       "success",
			version:    &outport.RegistryVersion{Version: "v1", SyncedAt: time.Now()},
			wantStatus: http.StatusOK,
		},
		{
			name:       "service error",
			err:        errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := inmocks.NewMockRegistryService(t)
			api := New(nil, nil, nil, nil, nil, service, nil, uuid.New())

			service.EXPECT().GetRegistryVersion(mock.Anything).Return(tc.version, tc.err).Once()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/registry/version", nil)
			api.Registry.GetRegistryVersion(c)
			require.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestRegistryAPIImportDataProblems(t *testing.T) {
	gin.SetMode(gin.TestMode)

	updatedAt := time.Date(2026, 6, 26, 0, 0, 0, 0, time.UTC)
	body, err := json.Marshal(dataProblemsImportRequest{
		Version:   "my-roadmap-v1",
		UpdatedAt: updatedAt,
		Problems: []domain.ProblemManifest{
			{
				Provider:      domain.ProviderLeetCode,
				ExternalID:    "1",
				Slug:          "two-sum",
				Title:         "Two Sum",
				Difficulty:    domain.DifficultyEasy,
				Tags:          []string{"array", "hash-table"},
				SourceURL:     "https://leetcode.com/problems/two-sum/",
				EstimatedTime: 15,
				ExecutionSpec: domain.ExecutionSpec{
					Mode:       domain.ExecutionModeFunction,
					Entrypoint: "twoSum",
				},
				TestCases: []domain.TestCaseManifest{
					{
						InputJSON:    []byte(`{"args":[[2,7],9]}`),
						ExpectedJSON: []byte(`[0,1]`),
					},
				},
			},
		},
		Manifests: []domain.ManifestRef{{Name: "my-roadmap", Path: "tracks/my-roadmap.json"}},
	})
	require.NoError(t, err)

	cases := []struct {
		name       string
		body       []byte
		imported   int
		err        error
		wantStatus int
		wantCall   bool
	}{
		{name: "bad request", body: []byte("{"), wantStatus: http.StatusBadRequest},
		{name: "success", body: body, imported: 1, wantStatus: http.StatusOK, wantCall: true},
		{
			name:       "service error",
			body:       body,
			err:        errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantCall:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := inmocks.NewMockRegistryService(t)
			api := New(nil, nil, nil, nil, nil, service, nil, uuid.New())
			if tc.wantCall {
				service.EXPECT().
					SyncRegistry(
						mock.Anything,
						"my-roadmap-v1",
						updatedAt,
						mock.MatchedBy(func(problems []domain.ProblemManifest) bool {
							return len(problems) == 1 &&
								problems[0].ExecutionSpec.Mode == domain.ExecutionModeFunction &&
								problems[0].ExecutionSpec.Entrypoint == "twoSum" &&
								string(problems[0].TestCases[0].InputJSON) == `{"args":[[2,7],9]}` &&
								string(problems[0].TestCases[0].ExpectedJSON) == `[0,1]`
						}),
						mock.Anything,
					).
					Return(tc.imported, tc.err).
					Once()
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/data/problems", bytes.NewReader(tc.body))
			c.Request.Header.Set("Content-Type", "application/json")
			api.Registry.ImportDataProblems(c)
			require.Equal(t, tc.wantStatus, w.Code)
		})
	}
}
