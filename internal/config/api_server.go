package config

type APIServer struct {
	DatabaseURL string `env:"DATABASE_URL" envDefault:"postgres://judgeloop:judgeloop@localhost:5432/judgeloop?sslmode=disable"`
	RedisURL    string `env:"REDIS_URL" envDefault:"localhost:6379"`
	Port        string `env:"PORT" envDefault:"8080"`
	UserID      string `env:"USER_ID" envDefault:"00000000-0000-0000-0000-000000000001"`
}

func LoadAPIServer() (APIServer, error) {
	cfg := APIServer{}
	return cfg, parse(&cfg)
}
