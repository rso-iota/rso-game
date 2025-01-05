package game

import (
	"net/http"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Player struct {
	info PlayerInfo

	lobby *Game

	conn *websocket.Conn

	send chan []byte
}

type PlayerInfo struct {
	Id        string `json:"sub"`
	Email     string `json:"email"`
	Username  string `json:"preferredUsername"`
	Surname   string `json:"familyName"`
	GivenName string `json:"givenName"`
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

func serveWebSocket(playerInfo PlayerInfo, game *Game, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WithError(err).Error("Failed to upgrade connection to websocket")
		return
	}

	player := &Player{
		info:  playerInfo,
		lobby: game,
		conn:  conn,
		send:  make(chan []byte),
	}

	game.connect <- player

	go player.receiveMessage()
	go player.sendMessage()
}
