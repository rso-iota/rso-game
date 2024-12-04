package main

import (
	"rso-game/game"
	"rso-game/server"
)

func main() {
	game.CreateGame() // Za testiranje
	game.CreateGame() // Za testiranje

	server.Start()
}
