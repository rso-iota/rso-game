package main

import (
	"rso-game/circuitbreaker"
	"rso-game/config"
	"rso-game/game"
	"rso-game/grpc"
	"rso-game/nats"
	"rso-game/server"
)

func main() {
	conf := config.Init()
	game.SetGlobalConfig(conf)

	circuitbreaker.InitBreakers()

	nats.Connect(conf.NatsURL)

	grpc.InitBotClient(conf.BotServiceURL)
	grpc.InitLobbyClient(conf.LobbyServiceURL)

	game.InitBackup(conf.BackupRedisUrl)
	game.RestoreFromBackup()
	game.EnsureGames(conf.NumTestGames)

	server.Start(conf)
}
