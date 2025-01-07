package grpc

import (
	"context"
	"os"
	"strings"
	"time"

	"rso-game/circuitbreaker"
	pb "rso-game/grpc/bots"

	log "github.com/sirupsen/logrus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

type BotClient struct {
	client pb.BotServiceClient
	conn   *grpc.ClientConn
}

var botClient *BotClient
var botsUrl string

func InitBotClient(address string) {
	botsUrl = address
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.WithError(err).Error("failed to connect to bot service")
	}

	client := pb.NewBotServiceClient(conn)
	botClient = &BotClient{
		client: client,
		conn:   conn,
	}

	log.Info("Connected to bot service at ", address)
}

func CreateBot(gameID string, botID string, token string, difficulty string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	hostname := os.Getenv("HOSTNAME")
	if strings.Contains(hostname, "statefulset") {
		splits := strings.Split(hostname, "-")
		podID := splits[len(splits)-1]
		hostname = "game-svc-" + podID
	}

	req := &pb.CreateBotRequest{
		BotId:       botID,
		AccessToken: token,
		Hostname:    hostname,
		Bot: &pb.Bot{
			GameId:     gameID,
			Difficulty: difficulty,
		},
	}

	defer func() error {
		if r := recover(); r != nil {
			return r.(error)
		}
		return nil
	}()

	_, err := circuitbreaker.BotsBreaker.Execute(func() (*pb.CreateBotResponse, error) {
		if botClient.conn == nil || botClient.conn.GetState() == connectivity.Shutdown {
			log.Info("Trying to reconnect to bot service")
			conn, err := grpc.NewClient(botsUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return nil, err
			}

			client := pb.NewBotServiceClient(conn)
			botClient = &BotClient{
				client: client,
				conn:   conn,
			}
			log.Info("Reconnected to bot service")
		}

		resp, err := botClient.client.CreateBot(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	})

	return err
}
