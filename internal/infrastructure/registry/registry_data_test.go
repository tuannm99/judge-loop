package registry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tuannm99/judge-loop/internal/domain"
)

func TestBundledJudgeReadyManifestsAreExecutable(t *testing.T) {
	workingDirectory, err := os.Getwd()
	require.NoError(t, err)
	registryPath := filepath.Clean(filepath.Join(workingDirectory, "..", "..", "..", "registry"))

	index, err := LoadIndex(registryPath)
	require.NoError(t, err)
	problems, err := LoadAllProblems(registryPath, index)
	require.NoError(t, err)

	judgeReady := 0
	for _, problem := range problems {
		if !problem.JudgeReady {
			continue
		}
		judgeReady++
		require.NoError(t, domain.ValidateJudgeReadyManifest(problem), problem.Slug)
	}
	require.Greater(t, judgeReady, 0)
}
