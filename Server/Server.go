package main

import (
	"context"
	"log"
	"net"
	"sync"

	"github.com/lottejd/MockExam/Increment"
	"google.golang.org/grpc"
)

const (
	port = ":8080"
)

var (
	inc  = -1
	lock sync.Mutex
)

type Server struct {
	Increment.UnimplementedIncrementServiceServer
}

func main() {

	grpcServer := grpc.NewServer()

	//setup listen on port
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// start the service / server on the specific port
	Increment.RegisterIncrementServiceServer(grpcServer, &Server{})
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server over port %s  %v", port, err)
	}

}

func (s *Server) Increment(ctx context.Context, incr *Increment.IncRequest) (*Increment.IncResponse, error) {
	lock.Lock()
	defer lock.Unlock()

	inc++

	return &Increment.IncResponse{Success: true, Inc: int32(inc)}, nil
}
