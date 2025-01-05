package config

import (
	"reflect"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	HttpPort       string `env:"GAME_HTTP_PORT"`
	GrpcPort       string `env:"GAME_GRPC_PORT"`
	NumTestGames   int    `env:"GAME_NUM_TEST_GAMES"`
	TestServer     bool   `env:"GAME_TEST_SERVER"`
	LogJSON        bool   `env:"GAME_LOG_JSON"`
	NatsURL        string `env:"GAME_NATS_URL"`
	AuthEndpoint   string `env:"GAME_AUTH_EP"`
	RequireAuth    bool   `env:"GAME_REQUIRE_AUTH"`
	BackupRedisUrl string `env:"GAME_BACKUP_REDIS_URL"`
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

	if config.LogJSON {
		log.SetFormatter(&log.JSONFormatter{})
	}
	log.SetLevel(log.DebugLevel)

	log.WithFields(fields).Info("Loaded config")

	return config
}
