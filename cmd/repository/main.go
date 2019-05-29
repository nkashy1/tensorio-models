package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/doc-ai/tensorio-models/authentication"
	"github.com/doc-ai/tensorio-models/server"
	"github.com/doc-ai/tensorio-models/storage"
	"github.com/doc-ai/tensorio-models/storage/gcs"
	"github.com/doc-ai/tensorio-models/storage/memory"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

func main() {
	/* BEGIN cli */
	// Backend specification
	Backends := map[string]func() storage.RepositoryStorage{
		"memory": memory.NewMemoryRepositoryStorage,
		"gcs":    gcs.GenerateNewGCSStorageFromEnv,
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

	tokenTypeToSet := &authentication.AuthenticationTokenTypeToSet{
		server.MODELS_ADMIN:  authentication.AuthenticationTokenSet{},
		server.MODELS_READER: authentication.AuthenticationTokenSet{},
		server.MODELS_WRITER: authentication.AuthenticationTokenSet{},
	}
	filePath := os.Getenv("AUTH_TOKENS_FILE")
	if filePath == "" {
		err := errors.New("AUTH_TOKENS_FILE must be defined")
		panic(err)
	}
	bucketName := repositoryBackend.GetBucketName()
	var auth authentication.Authenticator
	if bucketName == "" {
		auth = authentication.NewAuthenticator(&authentication.FileSystemAuthentication{
			TokenFilePath:  filePath,
			TokenTypeToSet: tokenTypeToSet,
		})
	} else {
		auth = authentication.NewAuthenticator(&authentication.GCSAuthentication{
			BucketName:     bucketName,
			TokenFilePath:  filePath,
			TokenTypeToSet: tokenTypeToSet,
		})
	}
	const grpcAddress = ":8080"
	const jsonRpcAddress = ":8081"
	server.StartGrpcAndProxyServer(repositoryBackend,
		grpcAddress, jsonRpcAddress, auth, make(chan string))
}
