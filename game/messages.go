package game

import "encoding/json"

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type JoinMessage struct {
	PlayerName string `json:"playerName"`
}

type MoveMessage struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

type SpawnMessage struct {
	Type string     `json:"type"`
	Data PlayerData `json:"data"`
}

type GameStateMessage struct {
	Type string    `json:"type"`
	Data GameState `json:"data"`
}

type PlayerLeftMessage struct {
	Type string     `json:"type"`
	Data PlayerData `json:"data"`
}
