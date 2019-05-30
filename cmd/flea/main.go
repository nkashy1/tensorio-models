package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/doc-ai/tensorio-models/authentication"
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

	modelsURI := os.Getenv("MODELS_URI")
	if modelsURI == "" {
		err := errors.New("MODELS_URI not set")
		panic(err)
	}
	tokenFilePath := os.Getenv("AUTH_TOKENS_FILE")
	if tokenFilePath == "" {
		err := errors.New("AUTH_TOKENS_FILE must be provided.")
		panic(err)
	}
	fleaBackend := backend(modelsURI)
	bucketName := fleaBackend.GetBucketName()
	var auth authentication.Authenticator
	// This is just a template that lists all tokens types that we care about.
	tokenTypeToSet := &authentication.AuthenticationTokenTypeToSet{
		flea_server.FleaAdmin:   authentication.AuthenticationTokenSet{},
		flea_server.FleaClient:  authentication.AuthenticationTokenSet{},
		flea_server.FleaTaskGen: authentication.AuthenticationTokenSet{},
	}
	if bucketName == "" {
		auth = authentication.NewAuthenticator(&authentication.FileSystemAuthentication{
			TokenFilePath:  tokenFilePath,
			TokenTypeToSet: tokenTypeToSet,
		})
	} else {
		auth = authentication.NewAuthenticator(&authentication.GCSAuthentication{
			BucketName:     bucketName,
			TokenFilePath:  tokenFilePath,
			TokenTypeToSet: tokenTypeToSet,
		})
	}
	const grpcAddress = ":8082"
	const jsonRpcAddress = ":8083"
	flea_server.StartGrpcAndProxyServer(fleaBackend,
		grpcAddress, jsonRpcAddress, auth, make(chan string))
}
