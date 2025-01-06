package server

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"rso-game/config"
	"rso-game/game"
	"strings"

	pb "rso-game/grpc/game"

	log "github.com/sirupsen/logrus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GrpcServer struct {
	pb.UnimplementedGameServiceServer
}

var hostname string

func (s *GrpcServer) CreateGame(_ context.Context, _ *pb.Empty) (*pb.GameLocation, error) {
	id := game.CreateGame()
	return &pb.GameLocation{Id: &pb.GameID{Id: id}, Hostname: hostname}, nil
}

func (s *GrpcServer) DeleteGame(_ context.Context, gameId *pb.GameID) (*pb.GameID, error) {
	id := gameId.GetId()
	err := game.DeleteGame(id)
	if err != nil {
		return nil, err
	}
	return &pb.GameID{Id: id}, nil
}

func (s *GrpcServer) LiveData(_ context.Context, gameId *pb.GameID) (*pb.GameData, error) {
	id := gameId.GetId()
	players, err := game.GetPlayerSizes(id)
	if err != nil {
		return nil, err
	}

	return players, nil
}

func serveHTTP(l net.Listener, config config.Config) {
	mux := http.NewServeMux()
	mux.HandleFunc("/connect/{gameID}", game.HandleNewConnection)

	log.Println("Starting HTTP server on " + l.Addr().String())

	// Self contained app - for testing
	if config.TestServer {
		log.Debug("Running in test mode, serving static files")
		mux.HandleFunc("/script.js", serveScript)
		mux.HandleFunc("/list", gameListHandler)
		mux.HandleFunc("/new", newGameHandler)
		mux.HandleFunc("/", serveStaticPage)
	}

	err := http.Serve(l, mux)
	if err != nil {
		log.Fatal(err)
	}
}

func serveGRPC(l net.Listener) {
	grpcServer := grpc.NewServer()
	pb.RegisterGameServiceServer(grpcServer, &GrpcServer{})

	reflection.Register(grpcServer)

	log.Println("Starting gRPC server on " + l.Addr().String())
	err := grpcServer.Serve(l)
	if err != nil {
		log.Fatal(err)
	}
}

func Start(config config.Config) {
	hostname = os.Getenv("HOSTNAME")
	if strings.Contains(hostname, "statefulset") {
		splits := strings.Split(hostname, "-")
		podID := splits[len(splits)-1]
		hostname = "game-svc-" + podID
	}

	httpListen, err := net.Listen("tcp", ":"+config.HttpPort)
	if err != nil {
		log.Fatal(err)
	}

	grpcListen, err := net.Listen("tcp", ":"+config.GrpcPort)
	if err != nil {
		log.Fatal(err)
	}

	go serveHTTP(httpListen, config)
	serveGRPC(grpcListen)
}

// These are the handlers for the self-contained app, not needed in the cluster
func serveStaticPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/index.html")
}

func serveScript(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/script.js")
}

func newGameHandler(w http.ResponseWriter, r *http.Request) {
	game := game.CreateGame()

	w.Write([]byte(game))
}

func gameListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(game.RunningGameIDs())
}
