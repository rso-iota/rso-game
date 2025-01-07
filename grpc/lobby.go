package grpc

import (
	"context"
	"rso-game/circuitbreaker"
	pb "rso-game/grpc/lobby"
	"time"

	log "github.com/sirupsen/logrus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

type LobbyClient struct {
	client pb.LobbyServiceClient
	conn   *grpc.ClientConn
}

var lobbyClient *LobbyClient
var lobbyUrl string

func InitLobbyClient(address string) {
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

func tryReconnect() error {
	log.Info("Trying to reconnect to bot service")
	conn, err := grpc.NewClient(lobbyUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	client := pb.NewLobbyServiceClient(conn)
	lobbyClient = &LobbyClient{
		client: client,
		conn:   conn,
	}
	log.Info("Reconnected to bot service")

	return nil
}

func NotifyGameDeleted(gameID string, reson pb.EndGameReason) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	req := &pb.EndGameRequest{
		Id: &pb.GameID{
			Id: gameID,
		},
		Reason: reson,
	}

	defer func() error {
		if r := recover(); r != nil {
			return r.(error)
		}
		return nil
	}()

	_, err := circuitbreaker.LobbyBreaker.Execute(func() (*pb.GameID, error) {
		if botClient.conn == nil || botClient.conn.GetState() == connectivity.Shutdown {
			err := tryReconnect()
			if err != nil {
				return nil, err
			}
		}

		resp, err := lobbyClient.client.DeleteGame(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	})

	return err
}

func NotifyLiveData(gameID string, players []*pb.Player) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	req := &pb.LiveDataRequest{
		Id: &pb.GameID{
			Id: gameID,
		},
		Players: players,
	}

	defer func() error {
		if r := recover(); r != nil {
			return r.(error)
		}
		return nil
	}()

	_, err := circuitbreaker.LobbyBreaker.Execute(func() (*pb.GameID, error) {
		if botClient.conn == nil || botClient.conn.GetState() == connectivity.Shutdown {
			if botClient.conn == nil || botClient.conn.GetState() == connectivity.Shutdown {
				err := tryReconnect()
				if err != nil {
					return nil, err
				}
			}
		}

		resp, err := lobbyClient.client.UpdateLiveData(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	})

	return err
}
