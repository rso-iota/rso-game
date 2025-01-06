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

func InitBackup(redisURL string) {
	client = redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: "",
		DB:       0,
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.WithError(err).Error("Failed to connect to Redis")
	} else {
		log.Info("Connected to Redis at ", redisURL)
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

func SaveToBackup(key string, data GameState) {
	err := client.Set(ctx, hostname+":"+key, ToBytes(data), 0).Err()
	if err != nil {
		log.WithError(err).Error("Failed to save game state")
	}
}

func LoadBackup() map[string]GameState {
	keys, err := client.Keys(ctx, hostname+":*").Result()
	if err != nil {
		log.WithError(err).Error("Failed to get keys")
	}

	games := make(map[string]GameState)

	for _, key := range keys {
		bytes, err := client.Get(ctx, key).Result()
		if err != nil {
			log.WithError(err).Error("Failed to get game state")
		}

		gameId := key[len(hostname)+1:]
		games[gameId] = FromBytes([]byte(bytes))
	}

	return games
}

func DeleteBackup(key string) {
	log.Info("Deleting backup for game ", key)
	err := client.Del(ctx, hostname+":"+key).Err()
	if err != nil {
		log.WithError(err).Error("Failed to delete game state")
	}
}
