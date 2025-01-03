package config

import (
	"reflect"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type Config struct {

	HTTPPort      string `env:"GAME_HTTP_PORT" envDefault:"8080"`
	GRPCPort      string `env:"GAME_GRPC_PORT" envDefault:"8081"`
	NumTestGames  int    `env:"GAME_NUM_TEST_GAMES" envDefault:"4"`
	TestServer    bool   `env:"GAME_TEST_SERVER" envDefault:"true"`
	LogJSON       bool   `env:"GAME_LOG_JSON" envDefault:"false"`
	BotServiceURL string `env:"GAME_BOT_SERVICE_URL" envDefault:"localhost:50051"`
	MinPlayers    int    `env:"GAME_MIN_PLAYERS" envDefault:"3"`
	NatsURL      string `env:"GAME_NATS_URL"`
	AuthEndpoint string `env:"GAME_AUTH_EP"`
	RequireAuth  bool   `env:"GAME_REQUIRE_AUTH"`
}

func Init() Config {
	godotenv.Load("defaults.env")

	var config Config
	err := env.Parse(&config)
	if err != nil {
		log.WithError(err).Fatal("Failed to parse config")
	}

	fields := log.Fields{}

	val := reflect.ValueOf(config)
	for i := 0; i < val.NumField(); i++ {
		fields[val.Type().Field(i).Name] = val.Field(i).Interface()
	}

	log.WithFields(fields).Info("Loaded config")

	if config.LogJSON {
		log.SetFormatter(&log.JSONFormatter{})
	}
	log.SetLevel(log.DebugLevel)

	return config
}
