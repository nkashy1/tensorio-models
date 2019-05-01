package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/doc-ai/tensorio-models/api"
	"github.com/doc-ai/tensorio-models/server"
	"github.com/doc-ai/tensorio-models/storage"
	"github.com/doc-ai/tensorio-models/storage/memory"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"strings"
)

func startGrpcServer(apiServer api.RepositoryServer) {
	serverAddress := ":8080"
	log.Println("Starting grpc on:", serverAddress)

	grpcServer := grpc.NewServer()
	lis, err := net.Listen("tcp", serverAddress)
	api.RegisterRepositoryServer(grpcServer, apiServer)
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
	/* BEGIN cli */
	// Backend specification
	Backends := map[string]func() storage.RepositoryStorage{
		"memory": memory.NewMemoryRepositoryStorage,
	}
	BackendKeys := make([]string, len(Backends))
	i := 0
	for k := range Backends {
		BackendKeys[i] = k
		i++
	}
	BackendChoices := strings.Join(BackendKeys, ",")

	var backendArg string
	backendUsage := fmt.Sprintf("Specifies the repository storage backend to be used; choices: %s", BackendChoices)
	flag.StringVar(&backendArg, "backend", "", backendUsage)

	flag.Parse()

	backend, exists := Backends[backendArg]
	if !exists {
		log.Fatalf("Unknown backend: %s. Choices are: %s", backendArg, BackendChoices)
	}
	/* END cli */

	repositoryBackend := backend()
	apiServer := server.NewServer(repositoryBackend)

	go startGrpcServer(apiServer)
	go startProxyServer()

	// sleep forever
	select {}
}
