package main

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/lottejd/MockExam/Increment"
	"google.golang.org/grpc"
)

const (
	port = 5000
)

var (
	inc  = -1
	lock sync.Mutex
)

type Server struct {
	Increment.UnimplementedIncrementServiceServer
	port        int
	latestValue int
	arbiter     sync.Mutex
}

func main() {
	serverPort := FindFreePort()
	if serverPort == -1 { // if no free port -1
		fmt.Printf("Can't start more than %v", 5)
		return
	}
	server := Server{port: serverPort, latestValue: inc, arbiter: sync.Mutex{}}
	fmt.Printf("Succesfully got port: %v\n", server.port) // sanity checks

	listen(&server)
	fmt.Println("main has ended")
}

func (s *Server) Increment(ctx context.Context, incr *Increment.IncRequest) (*Increment.IncResponse, error) {
	lock.Lock()
	defer lock.Unlock()

	inc++

	return &Increment.IncResponse{Success: true, Inc: int32(inc)}, nil
}

// connect to ports until a free port is found
func FindFreePort() int {
	for i := 1; i < 6; i++ {
		serverPort := port + i
		_, status := ConnectToPort(serverPort)
		if status == "alive" {
			continue
		} else {
			return serverPort
		}
	}
	return -1
}

// connect to a port and check if alive
func ConnectToPort(port int) (Increment.IncrementServiceClient, string) {
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(port), grpc.WithTimeout(time.Millisecond*250), grpc.WithInsecure()) // grpc.WithBlock(),
	if err == nil {
		ctx := context.Background()
		defer ctx.Done()
		client := Increment.NewIncrementServiceClient(conn)
		/* response, _ := client.JoinService(ctx, &Proto.JoinRequest{UserId: -1})
		return client, response.GetMsg() */
		return client, "alive"
	}
	return nil, "unknown"
}

func listen(s *Server) {

	//listen on port
	lis, err := net.Listen("tcp", "localhost:"+strconv.Itoa(s.port))
	if err != nil {
		fmt.Errorf(err.Error())
	}
	defer lis.Close()

	// register server this is a blocking call
	grpcServer := grpc.NewServer()
	Increment.RegisterIncrementServiceServer(grpcServer, s)
	errorMsg := grpcServer.Serve(lis)
	if errorMsg != nil {
		fmt.Errorf(errorMsg.Error())
	}
}
