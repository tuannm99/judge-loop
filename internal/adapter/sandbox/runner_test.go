package sandboxadapter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	outport "github.com/tuannm99/judge-loop/internal/port/out"
)

func TestNewRunner(t *testing.T) {
	runner := NewRunner()
	require.NotNil(t, runner)
}

func TestRunnerRunRejectsUnsupportedLanguage(t *testing.T) {
	runner := NewRunner()
	_, err := runner.Run(context.Background(), outport.RunRequest{
		Language: "ruby",
		Code:     "puts 1",
		Input:    "",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported language")
}
