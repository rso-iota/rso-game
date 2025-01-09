package server

import (
	"encoding/json"
	"net"
	"net/http"
	"os"
	"rso-game/config"
	"rso-game/game"
	"strings"

	log "github.com/sirupsen/logrus"
)

var podID string
var hostname string

func serveExternalHTTP(l net.Listener, config config.Config) {
	mux := http.NewServeMux()
	mux.HandleFunc("/connect/{gameID}", game.HandleNewConnection)

	log.Println("Starting external HTTP server on " + config.ExternalHttpPort)

	// Self contained app - for testing
	if config.TestServer {
		log.Debug("Running in test mode, serving static files")
		mux.HandleFunc("/script.js", serveScript)
		mux.HandleFunc("/list", gameListHandler)
		mux.HandleFunc("/", serveStaticPage)
	}

	err := http.Serve(l, mux)
	if err != nil {
		log.Fatal(err)
	}
}

func serverInternalHTTP(l net.Listener, config config.Config) {
	mux := http.NewServeMux()
	mux.HandleFunc("/game", gameCRUDhandler)

	log.Println("Starting internal HTTP server on " + config.InternalHttpPort)

	err := http.Serve(l, mux)
	if err != nil {
		log.Fatal(err)
	}
}

func Start(config config.Config) {
	hostname = os.Getenv("HOSTNAME")
	if strings.Contains(hostname, "statefulset") {
		splits := strings.Split(hostname, "-")
		podID = splits[len(splits)-1]
		hostname = "game-svc-" + podID
	}

	externalHttpListen, err := net.Listen("tcp", ":"+config.ExternalHttpPort)
	if err != nil {
		log.Fatal(err)
	}

	internalHttpListen, err := net.Listen("tcp", ":"+config.InternalHttpPort)
	if err != nil {
		log.Fatal(err)
	}

	go serveExternalHTTP(externalHttpListen, config)
	serverInternalHTTP(internalHttpListen, config)
}

// These are the handlers for the self-contained app, not needed in the cluster
func serveStaticPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/index.html")
}

func serveScript(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/script.js")
}

func newGameHandler(w http.ResponseWriter, _ *http.Request) {
	id := game.CreateGame()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": id, "serverId": podID})
}

func gameListHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(game.RunningGameIDs())
}

func deleteGameHandler(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("id")
	err := game.DeleteGame(gameID)
	if err != nil {
		log.WithError(err).Error("Failed to delete game over HTTP")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func gameCRUDhandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		newGameHandler(w, r)
	case http.MethodGet:
		gameListHandler(w, r)
	case http.MethodDelete:
		deleteGameHandler(w, r)
	}
}
