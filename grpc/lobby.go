package grpc

import (
	pb "rso-game/grpc/lobby"

	log "github.com/sirupsen/logrus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type LobbyClient struct {
	client pb.LobbyServiceClient
	conn   *grpc.ClientConn
}

var lobbyClient *LobbyClient
var lobbyUrl string

func InitBotLobby(address string) {
	lobbyUrl = address
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.WithError(err).Error("failed to connect to bot service")
	}

	client := pb.NewLobbyServiceClient(conn)
	lobbyClient = &LobbyClient{
		client: client,
		conn:   conn,
	}

	log.Info("Connected to bot service at ", address)
}

func Temp(gameID string, botID string, token string, difficulty string) error {
	return nil
}
