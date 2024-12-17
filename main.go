package main

import (
	"log"
	"rso-game/game"
	"rso-game/server"

	"github.com/spf13/viper"
)

func main() {
	viper.SetDefault("httpPort", "8080")
	viper.SetDefault("grpcPort", "8081")
	viper.SetDefault("numTestGames", 0)
	viper.SetDefault("testServer", false)

	viper.SetConfigFile("config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}

	testGames := viper.GetInt("numTestGames")
	if testGames > 0 {
		log.Println("Creating test games")
		for range viper.GetInt("numTestGames") {
			game.CreateGame()
		}
	}

	server.Start()
}
