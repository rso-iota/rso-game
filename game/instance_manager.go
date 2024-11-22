package game

import (
	"errors"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Lobby struct {
	ID      string
	players []*websocket.Conn
}

var runningLobbies = make(map[string]*Lobby)

func CreateLobby() *Lobby {
	id := uuid.New().String()
	Lobby := &Lobby{
		ID: id,
	}
	runningLobbies[id] = Lobby

	println("There are now", len(runningLobbies), "running game Lobbys")

	return Lobby
}

func RunningLobbyIDs() []string {
	ids := make([]string, len(runningLobbies))

	i := 0
	for id := range runningLobbies {
		ids[i] = id
		i++
	}

	return ids
}

func HandleConnection(conn *websocket.Conn) error {
	message := &Message{}

	err := conn.ReadJSON(message)
	if err != nil {
		return err
	}

	if message.Type == "join" {
		Lobby, ok := runningLobbies[string(message.Data)]
		if !ok {
			return errors.New("no such game Lobby")
		}

		Lobby.players = append(Lobby.players, conn)

		return nil
	}

	return errors.New("first message must be a join message")
}
