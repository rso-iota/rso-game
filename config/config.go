package config

import (
	"reflect"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	HttpPort       string `env:"HTTP_PORT"`
	GrpcPort       string `env:"GRPC_PORT"`
	NumTestGames   int    `env:"NUM_TEST_GAMES"`
	TestServer     bool   `env:"TEST_SERVER"`
	LogJSON        bool   `env:"LOG_JSON"`
	NatsURL        string `env:"NATS_URL"`
	AuthEndpoint   string `env:"AUTH_EP"`
	RequireAuth    bool   `env:"REQUIRE_AUTH"`
	BackupRedisUrl string `env:"BACKUP_REDIS_URL"`
	MinPlayers     int    `env:"MIN_PLAYERS"`
	BotServiceURL  string `env:"BOT_SERVICE_URL"`
	CreateGamePort string `env:"CREATE_GAME_HTTP_PORT"`
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
