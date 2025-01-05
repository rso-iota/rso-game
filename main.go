package main

import (
	"rso-game/config"
	"rso-game/game"
	"rso-game/nats"
	"rso-game/server"
)

func main() {
	conf := config.Init()
	game.SetGlobalConfig(conf)

	game.InitBackup(conf.BackupRedisUrl)
	nats.Connect(conf.NatsURL)

	game.RestoreFromBackup()
	game.EnsureGames(conf.NumTestGames)

	server.Start(conf)
}
