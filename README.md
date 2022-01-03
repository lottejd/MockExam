# How to run

- Start one front end

`go run .\FrontEnd\FrontEnd.go`

- start multiple servers (up to MAX_REPLICAS, default = 5)

`go run .\Server\Server.go`

- Start multiple clients

`go run .\Client\Client.go`

## Proto command

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative Election/Election.proto
