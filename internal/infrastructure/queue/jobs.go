// Package queue defines asynq job types and payload structures for
// async submission evaluation. The actual worker is implemented in
// apps/judge-worker (Milestone 6).
package queue

// TypeEvaluateSubmission is the asynq task type for evaluating a code submission.
const TypeEvaluateSubmission = "submission:evaluate"

// EvaluatePayload is the job payload for TypeEvaluateSubmission.
type EvaluatePayload struct {
	SubmissionID string `json:"submission_id"`
	UserID       string `json:"user_id"`
}
