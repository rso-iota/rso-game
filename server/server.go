package server

import (
	"encoding/json"
	"net/http"
	"rso-iota/game"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func serveStatic(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

	game.HandleConnection(conn)
}

func newGameHandler(w http.ResponseWriter, r *http.Request) {
	Lobby := game.CreateLobby()

	w.Write([]byte(Lobby.ID))
}

func gameListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(game.RunningLobbyIDs())
}

func Start() {
	http.HandleFunc("/list", gameListHandler)
	http.HandleFunc("/new", newGameHandler)
	http.HandleFunc("/connect", websocketHandler)
	http.HandleFunc("/", serveStatic)
	http.ListenAndServe(":8080", nil)
}
