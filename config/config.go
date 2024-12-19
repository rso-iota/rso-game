package config

type Config struct {
	HttpPort     string `env:"GAME_HTTP_PORT"`
	GrpcPort     string `env:"GAME_GRPC_PORT"`
	NumTestGames int    `env:"GAME_NUM_TEST_GAMES"`
	TestServer   bool   `env:"GAME_TEST_SERVER"`
	LogJSON      bool   `env:"GAME_LOG_JSON"`
}
