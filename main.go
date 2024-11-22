package main

import (
	"rso-iota/game"
	"rso-iota/server"
)

func main() {
	game.CreateLobby() // Za testiranje
	server.Start()
}
