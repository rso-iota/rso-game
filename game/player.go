package game

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Player struct {
	lobby *Game

	conn *websocket.Conn

	send chan []byte
}

func (p *Player) receiveMessage() {
	defer func() {
		p.lobby.disconnect <- p
		p.conn.Close()
	}()

	for {
		_, message, err := p.conn.ReadMessage()
		if err != nil {
			break
		}

		p.lobby.messages_in <- PlayerMessage{p, message}
	}
}

func (p *Player) sendMessage() {
	defer func() {
		p.conn.Close()
	}()

	for {
		message, ok := <-p.send
		if !ok {
			return
		}

		p.conn.WriteMessage(websocket.TextMessage, message)
	}
}

func serveWebSocket(game *Game, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.WithError(err).Error("Failed to upgrade connection to websocket")
		return
	}

	player := &Player{
		lobby: game,
		conn:  conn,
		send:  make(chan []byte),
	}

	game.connect <- player

	go player.receiveMessage()
	go player.sendMessage()
}
