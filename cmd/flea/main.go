package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/doc-ai/tensorio-models/flea_server"
	"github.com/doc-ai/tensorio-models/storage"
	"github.com/doc-ai/tensorio-models/storage/gcs"
	"github.com/doc-ai/tensorio-models/storage/memory"
	log "github.com/sirupsen/logrus"
)

func main() {
	/* BEGIN cli */
	// Backend specification
	Backends := map[string]func(string) storage.FleaStorage{
		"memory": memory.NewMemoryFleaStorage,
		"gcs":    gcs.GenerateNewFleaGCSStorageFromEnv,
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

	// For now assume Repository server is on the same machine.
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf(err.Error())
	}
	fleaBackend := backend("http://" + hostname + ":8081/v1/repository")
	const grpcAddress = ":8082"
	const jsonRpcAddress = ":8083"
	flea_server.StartGrpcAndProxyServer(fleaBackend,
		grpcAddress, jsonRpcAddress, make(chan string))
}
