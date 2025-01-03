package main

import (
	"reflect"
	"rso-game/config"
	"rso-game/game"
	"rso-game/server"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func main() {
	godotenv.Load("defaults.env")

	var config config.Config
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

	// Initialize bot client
	botClient, err := game.NewBotClient(config.BotServiceURL)
	if err != nil {
		log.WithError(err).Fatal("Failed to create bot client")
	}
	defer botClient.Close()

	// Create test games if configured
	testGames := config.NumTestGames
	if testGames > 0 {
		log.Debug("Creating test games")
		for range testGames {
			game.CreateGame(botClient, config.MinPlayers)
		}
	}

	server.Start(config)
}