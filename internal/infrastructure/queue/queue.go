// Package queue provides asynq client and server helpers for async submission evaluation.
package queue

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hibiken/asynq"
)

// parseRedisAddr normalises a Redis URL to the "host:port" form asynq expects.
// Accepts both "redis://host:port" and bare "host:port".
func parseRedisAddr(redisURL string) string {
	if strings.HasPrefix(redisURL, "redis://") {
		return strings.TrimPrefix(redisURL, "redis://")
	}
	return redisURL
}

// NewClient returns an asynq client connected to the given Redis address.
func NewClient(redisURL string) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{Addr: parseRedisAddr(redisURL)})
}

// NewServer returns an asynq server connected to the given Redis address.
func NewServer(redisURL string, concurrency int) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{Addr: parseRedisAddr(redisURL)},
		asynq.Config{Concurrency: concurrency},
	)
}

// NewEvaluateTask creates an asynq task carrying an EvaluatePayload.
func NewEvaluateTask(payload EvaluatePayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	return asynq.NewTask(TypeEvaluateSubmission, data), nil
}
