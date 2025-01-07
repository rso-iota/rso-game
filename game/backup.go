package game

import (
	"bytes"
	"context"
	"encoding/gob"
	"os"
	"rso-game/circuitbreaker"

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
	err := breakerSetState(hostname+":"+key, ToBytes(data))
	if err != nil {
		log.WithError(err).Error("Failed to save game state")
	}
}

func LoadBackup() map[string]GameState {
	keys, err := breakerGetKeys(hostname + ":*")
	if err != nil {
		log.WithError(err).Error("Failed to get keys")
	}

	games := make(map[string]GameState)

	for _, key := range keys {
		bytes, err := breakerGetState(key)
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
	err := breakerDeleteBackup(hostname + ":" + key)
	if err != nil {
		log.WithError(err).Error("Failed to delete game state")
	}
}

func breakerGetKeys(pattern string) ([]string, error) {
	defer func() error {
		if r := recover(); r != nil {
			return r.(error)
		}
		return nil
	}()

	keys, err := circuitbreaker.RedisBreaker.Execute(func() (interface{}, error) {
		keys, err := client.Keys(ctx, pattern).Result()
		return keys, err
	})

	return keys.([]string), err
}

func breakerSetState(key string, state []byte) error {
	defer func() error {
		if r := recover(); r != nil {
			return r.(error)
		}
		return nil
	}()

	_, err := circuitbreaker.RedisBreaker.Execute(func() (interface{}, error) {
		err := client.Set(ctx, key, state, 0).Err()
		return nil, err
	})

	return err
}

func breakerDeleteBackup(key string) error {
	defer func() error {
		if r := recover(); r != nil {
			return r.(error)
		}
		return nil
	}()

	_, err := circuitbreaker.RedisBreaker.Execute(func() (interface{}, error) {
		err := client.Del(ctx, key).Err()
		return nil, err
	})

	return err
}

func breakerGetState(key string) (string, error) {
	defer func() error {
		if r := recover(); r != nil {
			return r.(error)
		}
		return nil
	}()

	state, err := circuitbreaker.RedisBreaker.Execute(func() (interface{}, error) {
		state, err := client.Get(ctx, key).Result()
		return state, err
	})

	return state.(string), err
}
