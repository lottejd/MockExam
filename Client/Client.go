package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/lottejd/MockExam/Increment"
	"google.golang.org/grpc"
)

const (
	address = "localhost:8080"
)

func main() {
	// init
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// create client
	inc := Increment.NewIncrementServiceClient(conn)
	ctx := context.Background()

	for {
		// to ensure "enter" has been hit before publishing - skud ud til Mie
		reader, err := bufio.NewReader(os.Stdin).ReadString('\n')
		// remove newline windows format "\r\n"
		input := strings.TrimSuffix(reader, "\r\n")
		if err != nil {
			fmt.Errorf("bad bufio input")
		}
		if len(input) > 0 {
			Inc(ctx, inc)
		}
	}
}

func Inc(ctx context.Context, inc Increment.IncrementServiceClient) {
	response, err := inc.Increment(ctx, &Increment.IncRequest{})
	if err != nil {
		fmt.Errorf("increment error")
	}
	if response.Success {
		fmt.Println("Increment was successful, value is: ", response.Inc)
	}
}
