package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/doc-ai/tensorio-models/server"
	"github.com/doc-ai/tensorio-models/storage"
	"github.com/doc-ai/tensorio-models/storage/memory"
	log "github.com/sirupsen/logrus"
)

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
	const grpcAddress = ":8080"
	const jsonRpcAddress = ":8081"
	server.StartGrpcAndProxyServer(repositoryBackend,
		grpcAddress, jsonRpcAddress, make(chan string))
}
