package config

type JudgeWorker struct {
	DatabaseURL   string `env:"DATABASE_URL"    envDefault:"postgres://judgeloop:judgeloop@localhost:5432/judgeloop"`
	Concurrency   int    `env:"CONCURRENCY"     envDefault:"2"`
	TimeLimitSecs int    `env:"TIME_LIMIT_SECS" envDefault:"10"`
	WorkerID      string `env:"WORKER_ID"       envDefault:"judge-worker"`
}

func LoadJudgeWorker() (JudgeWorker, error) {
	cfg := JudgeWorker{}
	return cfg, parse(&cfg)
}
