package game

import (
	"context"
	"fmt"
	"time"

	pb "rso-game/grpc/bot"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type BotClient struct {
	client pb.BotServiceClient
	conn   *grpc.ClientConn
}

func NewBotClient(address string) (*BotClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to bot service: %v", err)
	}

	client := pb.NewBotServiceClient(conn)
	return &BotClient{
		client: client,
		conn:   conn,
	}, nil
}

func (bc *BotClient) Close() error {
	return bc.conn.Close()
}

func (bc *BotClient) CreateBot(gameID string, difficulty string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	botID := uuid.New().String()
	req := &pb.CreateBotRequest{
		BotId: botID,
		Bot: &pb.Bot{
			GameId:     gameID,
			Difficulty: difficulty,
		},
	}

	resp, err := bc.client.CreateBot(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create bot: %v", err)
	}

	return resp.BotId, nil
}
