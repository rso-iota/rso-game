package main

import (
	"rso-iota/game"
	"rso-iota/server"
)

func main() {
	l := game.CreateGame() // Za testiranje
	go l.Run()

	l = game.CreateGame() // Za testiranje
	go l.Run()

	server.Start()
}
