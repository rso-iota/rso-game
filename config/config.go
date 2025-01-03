package config

type Config struct {
	HTTPPort      string `env:"GAME_HTTP_PORT" envDefault:"8080"`
	GRPCPort      string `env:"GAME_GRPC_PORT" envDefault:"8081"`
	NumTestGames  int    `env:"GAME_NUM_TEST_GAMES" envDefault:"4"`
	TestServer    bool   `env:"GAME_TEST_SERVER" envDefault:"true"`
	LogJSON       bool   `env:"GAME_LOG_JSON" envDefault:"false"`
	BotServiceURL string `env:"GAME_BOT_SERVICE_URL" envDefault:"localhost:50051"`
	MinPlayers    int    `env:"GAME_MIN_PLAYERS" envDefault:"3"`
}
