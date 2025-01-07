package main

import (
	"rso-game/bots"
	"rso-game/circuitbreaker"
	"rso-game/config"
	"rso-game/game"
	"rso-game/nats"
	"rso-game/server"
)

func main() {
	conf := config.Init()
	game.SetGlobalConfig(conf)

	circuitbreaker.InitBreakers()

	nats.Connect(conf.NatsURL)
	bots.InitBotClient(conf.BotServiceURL)

	game.InitBackup(conf.BackupRedisUrl)
	game.RestoreFromBackup()
	game.EnsureGames(conf.NumTestGames)

	server.Start(conf)
}
