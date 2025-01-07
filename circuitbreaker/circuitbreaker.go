package circuitbreaker

import (
	botPb "rso-game/grpc/bots"
	lobbyPb "rso-game/grpc/lobby"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/sony/gobreaker/v2"
)

var BotsBreaker *gobreaker.CircuitBreaker[*botPb.CreateBotResponse]
var LobbyBreaker *gobreaker.CircuitBreaker[*lobbyPb.GameID]
var NatsBreaker *gobreaker.CircuitBreaker[interface{}]
var RedisBreaker *gobreaker.CircuitBreaker[interface{}]

// var redisBreaker *gobreaker.CircuitBreaker

func onChange(name string, from gobreaker.State, to gobreaker.State) {
	if to == gobreaker.StateOpen {
		log.WithField("type", "breaker").Error(name + " breaker is open")
	} else if to == gobreaker.StateHalfOpen {
		log.WithField("type", "breaker").Warn(name + " breaker is half open")
	} else if to == gobreaker.StateClosed {
		log.WithField("type", "breaker").Info(name + " breaker is closed")
	}
}

func InitBreakers() {
	BotsBreaker = gobreaker.NewCircuitBreaker[*botPb.CreateBotResponse](gobreaker.Settings{
		Name:          "grpcBreaker",
		Timeout:       5 * time.Second,
		OnStateChange: onChange,
	})

	NatsBreaker = gobreaker.NewCircuitBreaker[interface{}](gobreaker.Settings{
		Name:          "natsBreaker",
		Timeout:       5 * time.Second,
		OnStateChange: onChange,
	})

	RedisBreaker = gobreaker.NewCircuitBreaker[interface{}](gobreaker.Settings{
		Name:          "redisBreaker",
		Timeout:       5 * time.Second,
		OnStateChange: onChange,
	})

	LobbyBreaker = gobreaker.NewCircuitBreaker[*lobbyPb.GameID](gobreaker.Settings{
		Name:          "lobbyBreaker",
		Timeout:       5 * time.Second,
		OnStateChange: onChange,
	})

	log.Info("Circuit breakers initialized")
}
