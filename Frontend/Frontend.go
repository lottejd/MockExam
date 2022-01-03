package main

import (
	"context"
	"errors"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/lottejd/MockExam/Increment"
	"google.golang.org/grpc"
)

const (
	SERVER_PORT       = 5001
	FRONT_END_ADDRESS = "localhost:5000"
)

type Value struct {
	Value int32
	//UserId int32
}

type FrontEnd struct {
	Increment.UnimplementedIncrementServiceServer
	replicaServerPorts map[int]bool
	arbiter            sync.Mutex
	Value              int32
}

func main() {

	//init
	replicaServers := make(map[int]bool)

	frontEnd := FrontEnd{replicaServerPorts: replicaServers, arbiter: sync.Mutex{}}
	go listen(&frontEnd)
	for {
		// begin searching for replicas/servers every 5 seconds
		frontEnd.FindActiveServers()
		time.Sleep(5 * time.Second)
	}
}

func (fe *FrontEnd) Increment(ctx context.Context, request *Increment.IncRequest) (*Increment.IncResponse, error) {
	values := make(map[Value]int)
	fe.arbiter.Lock()
	for port, alive := range fe.replicaServerPorts {
		if alive {
			client, state := ConnectToPort(port)
			if state == "alive" {
				response, _ := client.Increment(ctx, &Increment.IncRequest{})
				value := Value{Value: int32(response.Inc)}
				temp := values[value]
				values[value] = (temp + 1)
			}
		}
	}
	fe.arbiter.Unlock()
	currentReplicas := 0
	for _, votes := range values {
		currentReplicas += votes
	}
	for value, votes := range values {
		if votes > currentReplicas/2 {
			return &Increment.IncResponse{Inc: value.Value}, nil
		}
	}

	return nil, errors.New("replicas couldn't agree on one value")
}

// check if a replica/server is running on the port, indicate as alive by setting map to true, checking ports up to MAX_REPLICAS
func (fe *FrontEnd) FindActiveServers() {
	for i := 0; i < 10; i++ {
		serverPort := SERVER_PORT + i
		_, status := ConnectToPort(serverPort)
		if status == "alive" {
			// fmt.Printf("found alive server at port : %v\n", serverPort)
			fe.replicaServerPorts[serverPort] = true
		} else if status == "unknown" {
			fe.replicaServerPorts[serverPort] = false
		}
	}
}

// start front end service
func listen(fe *FrontEnd) {

	//listen on port
	lis, err := net.Listen("tcp", FRONT_END_ADDRESS)
	CheckError(err, "server setup net.listen")
	defer lis.Close()

	// register server this is a blocking call
	grpcServer := grpc.NewServer()
	Increment.RegisterIncrementServiceServer(grpcServer, fe)
	errorMsg := grpcServer.Serve(lis)
	CheckError(errorMsg, "server listen register server service")
}

func CheckError(err error, msg string) {
	if err != nil {
		log.Fatalf("happened inside method: %s err: %v", msg, err)
	}
}

// connect to a port and check if alive
func ConnectToPort(port int) (Increment.IncrementServiceClient, string) {
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(port), grpc.WithTimeout(time.Millisecond*250), grpc.WithInsecure()) // grpc.WithBlock(),
	if err == nil {
		ctx := context.Background()
		defer ctx.Done()
		client := Increment.NewIncrementServiceClient(conn)
		return client, "success"
	}
	return nil, "unknown"
}
