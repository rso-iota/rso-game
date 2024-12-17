package main

import (
	"rso-game/game"
	"rso-game/server"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	viper.SetDefault("httpPort", "8080")
	viper.SetDefault("grpcPort", "8081")
	viper.SetDefault("numTestGames", 0)
	viper.SetDefault("testServer", false)
	viper.SetDefault("logJSON", false)

	viper.SetConfigFile("config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}

	if viper.GetBool("logJSON") {
		log.SetFormatter(&log.JSONFormatter{})
	}
	log.SetLevel(log.DebugLevel)

	testGames := viper.GetInt("numTestGames")
	if testGames > 0 {
		log.Debug("Creating test games")
		for range viper.GetInt("numTestGames") {
			game.CreateGame()
		}
	}

	server.Start()
}
