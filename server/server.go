package server

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"rso-game/game"

	pb "rso-game/grpc"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GrpcServer struct {
	pb.UnimplementedGameServiceServer
}

func (s *GrpcServer) CreateGame(_ context.Context, _ *pb.Empty) (*pb.GameID, error) {
	id := game.CreateGame()
	return &pb.GameID{Id: id}, nil
}

func (s *GrpcServer) ListRunningGames(_ context.Context, _ *pb.Empty) (*pb.GameIDList, error) {
	ids := game.RunningGameIDs()
	return &pb.GameIDList{Ids: ids}, nil
}

func serveHTTP(l net.Listener) {
	mux := http.NewServeMux()
	mux.HandleFunc("/connect/{gameID}", game.HandleNewConnection)

	log.Println("Starting HTTP server on port " + viper.GetString("httpPort"))

	// Self contained app - for testing
	if viper.GetBool("testServer") {
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

	log.Println("Starting gRPC server on port " + viper.GetString("grpcPort"))
	err := grpcServer.Serve(l)
	if err != nil {
		log.Fatal(err)
	}
}

func Start() {
	httpListen, err := net.Listen("tcp", ":"+viper.GetString("httpPort"))
	if err != nil {
		log.Fatal(err)
	}

	grpcListen, err := net.Listen("tcp", ":"+viper.GetString("grpcPort"))
	if err != nil {
		log.Fatal(err)
	}

	go serveHTTP(httpListen)
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
