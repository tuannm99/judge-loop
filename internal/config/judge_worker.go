package config

type JudgeWorker struct {
	DatabaseURL   string `env:"DATABASE_URL" envDefault:"postgres://judgeloop:judgeloop@localhost:5432/judgeloop?sslmode=disable"`
	RedisURL      string `env:"REDIS_URL" envDefault:"localhost:6379"`
	Concurrency   int    `env:"CONCURRENCY" envDefault:"2"`
	TimeLimitSecs int    `env:"TIME_LIMIT_SECS" envDefault:"10"`
}

func LoadJudgeWorker() (JudgeWorker, error) {
	cfg := JudgeWorker{}
	return cfg, parse(&cfg)
}
