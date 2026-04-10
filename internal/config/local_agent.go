package config

type LocalAgent struct {
	ServerURL    string `env:"JUDGE_SERVER_URL" envDefault:"http://localhost:8080"`
	BindAddress  string `env:"JUDGE_BIND_ADDRESS" envDefault:"127.0.0.1"`
	Port         int    `env:"JUDGE_PORT" envDefault:"7070"`
	UserID       string `env:"JUDGE_USER_ID" envDefault:"00000000-0000-0000-0000-000000000001"`
	DataDir      string `env:"JUDGE_DATA_DIR"`
	RegistryPath string `env:"JUDGE_REGISTRY_PATH" envDefault:"./registry"`
}

func LoadLocalAgent() (LocalAgent, error) {
	cfg := LocalAgent{}
	if err := parse(&cfg); err != nil {
		return LocalAgent{}, err
	}
	if cfg.DataDir == "" {
		cfg.DataDir = defaultDataDir()
	}
	return cfg, nil
}
