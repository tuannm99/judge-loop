package out

type EvaluationPublisher interface {
	PublishEvaluation(job EvaluateSubmissionJob) error
}
