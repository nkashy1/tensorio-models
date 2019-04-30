package main

import (
	"context"
	"github.com/doc-ai/tensorio-models/api"
	"github.com/doc-ai/tensorio-models/server"
	"github.com/doc-ai/tensorio-models/storage/memory"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"time"
)

func startGrpcServer() {
	serverAddress := ":8080"
	log.Println("Starting grpc on:", serverAddress)
	var srv api.RepositoryServer
	srv = server.NewServer(memory.NewMemoryRepositoryStorage())

	grpcServer := grpc.NewServer()
	lis, err := net.Listen("tcp", serverAddress)
	api.RegisterRepositoryServer(grpcServer, srv)
	if err != nil {
		log.Fatalln(err)
	}

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("over")
}

func startProxyServer() {
	grpcServerAddress := ":8080"
	serverAddress := ":8081"
	log.Println("Starting json-rpc on:", serverAddress)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := api.RegisterRepositoryHandlerFromEndpoint(ctx, mux, grpcServerAddress, opts)
	if err != nil {
		log.Fatalln(err)
	}

	err = http.ListenAndServe(serverAddress, mux)
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	go startGrpcServer()
	go startProxyServer()
	time.Sleep(1 * time.Hour)
}
