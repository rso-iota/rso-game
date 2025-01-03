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
	game.SetConfig(&conf)

	log.SetLevel(log.DebugLevel)

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
