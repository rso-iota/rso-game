package server

import (
	"encoding/json"
	"net/http"
	"rso-game/game"
)

func serveStaticPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

// Janky jank
func serveScript(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "script.js")
}

func newGameHandler(w http.ResponseWriter, r *http.Request) {
	game := game.CreateGame()

	w.Write([]byte(game.ID))
}

func gameListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(game.RunningGameIDs())
}

func Start() {
	http.HandleFunc("/list", gameListHandler)
	http.HandleFunc("/new", newGameHandler)
	http.HandleFunc("/connect/{gameID}", game.HandleNewConnection)
	http.HandleFunc("/script.js", serveScript)
	http.HandleFunc("/", serveStaticPage)

	http.ListenAndServe(":8080", nil)
}
