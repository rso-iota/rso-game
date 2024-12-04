package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"rso-game/game"

	pb "rso-game/grpc"

	"github.com/soheilhy/cmux"
	"golang.org/x/sync/errgroup"
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

func serveHTTP(l net.Listener) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/connect/{gameID}", game.HandleNewConnection)

	// Self contained app - for testing
	if os.Getenv("DEV") == "true" {
		fmt.Println("Using the testing server")
		mux.HandleFunc("/script.js", serveScript)
		mux.HandleFunc("/list", gameListHandler)
		mux.HandleFunc("/new", newGameHandler)
		mux.HandleFunc("/", serveStaticPage)
	}

	return http.Serve(l, mux)
}

func serveGRPC(l net.Listener) error {
	grpcServer := grpc.NewServer()
	pb.RegisterGameServiceServer(grpcServer, &GrpcServer{})

	reflection.Register(grpcServer)

	return grpcServer.Serve(l)
}

func Start() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	m := cmux.New(lis)
	grpcL := m.Match(cmux.HTTP2())
	httpL := m.Match(cmux.HTTP1Fast())

	g := new(errgroup.Group)
	g.Go(func() error { return serveGRPC(grpcL) })
	g.Go(func() error { return serveHTTP(httpL) })
	g.Go(func() error { return m.Serve() })

	log.Println("run server:", g.Wait())
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
