package game

import (
	"bytes"
	"context"
	"encoding/gob"
	"os"

	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

var ctx = context.Background()
var client *redis.Client
var hostname string = os.Getenv("HOSTNAME")
var url string

func InitBackup(redisURL string) {
	url = redisURL
	client = redis.NewClient(&redis.Options{
		Addr:     url,
		Password: "",
		DB:       0,
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.WithError(err).Error("Failed to connect to Redis")
	} else {
		log.Info("Connected to Redis at ", url)
	}
}

func ToBytes(data GameState) []byte {
	var bytes bytes.Buffer
	enc := gob.NewEncoder(&bytes)

	err := enc.Encode(data)
	if err != nil {
		log.WithError(err).Error("Failed to encode game state")
	}

	return bytes.Bytes()
}

func FromBytes(data []byte) GameState {
	var game GameState
	dec := gob.NewDecoder(bytes.NewReader(data))

	err := dec.Decode(&game)
	if err != nil {
		log.WithError(err).Error("Failed to decode game state")
	}

	return game
}

func dbError(msg string, err error) {
	log.WithError(err).Error(msg)
	log.Debug("Trying to reconnect to Redis")
	if client.Ping(ctx).Err() != nil {
		InitBackup(url)
	}
}

func SaveToBackup(key string, data GameState) {
	err := client.Set(ctx, hostname+":"+key, ToBytes(data), 0).Err()
	if err != nil {
		dbError("Failed to save game state", err)
	}
}

func LoadBackup() map[string]GameState {
	keys, err := client.Keys(ctx, hostname+":*").Result()
	if err != nil {
		dbError("Failed to get keys", err)
	}

	games := make(map[string]GameState)

	for _, key := range keys {
		bytes, err := client.Get(ctx, key).Result()
		if err != nil {
			dbError("Failed to get game state", err)
		}

		gameId := key[len(hostname)+1:]
		games[gameId] = FromBytes([]byte(bytes))
	}

	return games
}
