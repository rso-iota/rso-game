package main

import (
	"rso-game/game"
	"rso-game/server"
)

func main() {
	l := game.CreateGame() // Za testiranje
	go l.Run()

	l = game.CreateGame() // Za testiranje
	go l.Run()

	server.Start()
}
