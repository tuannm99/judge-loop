package queue

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

func TestParseRedisAddr(t *testing.T) {
	require.Equal(t, "localhost:6379", parseRedisAddr("redis://localhost:6379"))
	require.Equal(t, "localhost:6379", parseRedisAddr("localhost:6379"))
}

func TestNewClientAndServer(t *testing.T) {
	require.NotNil(t, NewClient("redis://localhost:6379"))
	require.NotNil(t, NewServer("localhost:6379", 2))
}

func TestNewEvaluateTask(t *testing.T) {
	payload := EvaluatePayload{SubmissionID: uuid.NewString(), UserID: uuid.NewString()}
	task, err := NewEvaluateTask(payload)
	require.NoError(t, err)
	require.Equal(t, TypeEvaluateSubmission, task.Type())

	var got EvaluatePayload
	require.NoError(t, json.Unmarshal(task.Payload(), &got))
	require.Equal(t, payload, got)

	require.IsType(t, &asynq.Task{}, task)
}
