package main

import (
	"rso-game/config"
	"rso-game/game"
	"rso-game/nats"
	"rso-game/server"

	log "github.com/sirupsen/logrus"
)

func main() {
	conf := config.Init()


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

	game.SetConfig(&config)

	game.SetGlobalConfig(conf)

	testGames := conf.NumTestGames
	if testGames > 0 {
		log.Debug("Creating test games")
		for range testGames {
			game.CreateGame()
		}
	}

	nats.Connect(conf.NatsURL)
	server.Start(conf)
}
