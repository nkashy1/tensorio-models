package flea_server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"

	"github.com/doc-ai/tensorio-models/api"
	"github.com/doc-ai/tensorio-models/authentication"
	"github.com/doc-ai/tensorio-models/common"
	"github.com/doc-ai/tensorio-models/storage"
	"github.com/doc-ai/tensorio-models/storage/gcs"
	"github.com/doc-ai/tensorio-models/storage/memory"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type flea_server struct {
	storage       storage.FleaStorage
	authenticator authentication.Authenticator
}

const (
	FLEA_ADMIN    authentication.AuthenticationTokenType = "FleaAdmin"
	FLEA_CLIENT   authentication.AuthenticationTokenType = "FleaClient"
	FLEA_TASK_GEN authentication.AuthenticationTokenType = "FleaTaskGen"
)

// NewServer - Creates an api.RepositoryServer which handles gRPC requests using a given
// storage.RepositoryStorage backend
func NewServer(storage storage.FleaStorage) api.FleaServer {
	tokenFilePath := os.Getenv("AUTH_TOKENS_FILE")
	if tokenFilePath == "" {
		err := errors.New("AUTH_TOKENS_FILE must be provided.")
		panic(err)
	}

	tokenTypeToSet := &authentication.AuthenticationTokenTypeToSet{
		FLEA_ADMIN:    authentication.AuthenticationTokenSet{},
		FLEA_CLIENT:   authentication.AuthenticationTokenSet{},
		FLEA_TASK_GEN: authentication.AuthenticationTokenSet{},
	}
	var auth authentication.Authenticator
	bucketName := storage.GetBucketName()
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
	return &flea_server{
		storage:       storage,
		authenticator: auth,
	}

}

func startGrpcServer(apiServer api.FleaServer, serverAddress string) {
	log.Println("Starting FLEA gRPC on:", serverAddress)

	grpcServer := grpc.NewServer()
	lis, err := net.Listen("tcp", serverAddress)
	if err != nil {
		log.Fatalln(err)
	}

	api.RegisterFleaServer(grpcServer, apiServer)

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("over")
}

func startProxyServer(grpcServerAddress string, jsonServerAddress string) {
	log.Println("Starting json-rpc on:", jsonServerAddress)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// Make the JSON output print default values.
	mux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}))
	opts := []grpc.DialOption{grpc.WithInsecure()}
	// Note the *Flea* handler
	err := api.RegisterFleaHandlerFromEndpoint(ctx, mux, grpcServerAddress, opts)
	if err != nil {
		log.Fatalln(err)
	}

	err = http.ListenAndServe(jsonServerAddress, mux)
	if err != nil {
		log.Fatalln(err)
	}
}

// StartGrpcAndProxyServer - Given a repository storage backend, this function starts a
// new gRPC and JSON-RPC server in separate threads and waits until a message is received on the stopRequested channel.
func StartGrpcAndProxyServer(storage storage.FleaStorage,
	grpcServerAddress string, jsonServerAddress string,
	stopRequested <-chan string) {
	apiServer := NewServer(storage)
	go startGrpcServer(apiServer, grpcServerAddress)
	go startProxyServer(grpcServerAddress, jsonServerAddress)
	stopReason := <-stopRequested
	log.Println("Stopping server due to:", stopReason)
}

func (srv *flea_server) Healthz(ctx context.Context, req *api.HealthCheckRequest) (*api.HealthCheckResponse, error) {
	log.Println("Health Request")
	resp := &api.HealthCheckResponse{
		Status: api.HealthCheckResponse_SERVING,
	}
	return resp, nil
}

func (srv *flea_server) Config(ctx context.Context, req *api.ConfigRequest) (*api.ConfigResponse, error) {
	log.Println("Config Request")
	storageType := srv.storage.GetStorageType()
	storageTypeEnum := api.ConfigResponse_INVALID
	switch storageType {
	case memory.StorageType:
		storageTypeEnum = api.ConfigResponse_MEMORY
	case gcs.StorageType:
		storageTypeEnum = api.ConfigResponse_GOOGLE_CLOUD_STORAGE
	}
	resp := &api.ConfigResponse{
		BackendType: storageTypeEnum,
	}
	return resp, nil
}

func (srv *flea_server) Admin(ctx context.Context, req *api.AdminRequest) (*api.AdminResponse, error) {
	auth_err := srv.authenticator.CheckAuthentication(ctx, FLEA_ADMIN)
	if auth_err != nil {
		return nil, auth_err
	}
	if req.Type != api.AdminRequest_RELOAD_TOKENS {
		return nil, errors.New("Unknown admin request type!")
	}
	err := srv.authenticator.ReloadAuthenticationTokens(ctx)
	if err != nil {
		return nil, err
	}
	return &api.AdminResponse{Info: "Updated Authentication Tokens"}, nil
}

func (srv *flea_server) CreateTask(ctx context.Context, req *api.TaskDetails) (*api.TaskDetails, error) {
	auth_err := srv.authenticator.CheckAuthentication(ctx, FLEA_TASK_GEN)
	if auth_err != nil {
		return nil, auth_err
	}
	log.Println("CreateTask:", req)
	if req.ModelId == "" {
		return nil, storage.ErrMissingModelId
	}
	if !common.IsValidID(req.ModelId) {
		return nil, storage.ErrInvalidModelId
	}
	if req.HyperparametersId == "" {
		return nil, storage.ErrMissingHyperparametersId
	}
	if !common.IsValidID(req.HyperparametersId) {
		return nil, storage.ErrInvalidHyperparametersId
	}
	if req.CheckpointId == "" {
		return nil, storage.ErrMissingCheckpointId
	}
	if !common.IsValidID(req.CheckpointId) {
		return nil, storage.ErrInvalidCheckpointId
	}
	if req.TaskId == "" {
		return nil, storage.ErrMissingTaskId
	}
	if !common.IsValidID(req.TaskId) {
		return nil, storage.ErrInvalidTaskId
	}
	err := srv.storage.AddTask(ctx, *req)
	if err != nil {
		return nil, err
	}
	resp, err := srv.storage.GetTask(ctx, req.TaskId)
	return &resp, err
}

func (srv *flea_server) ModifyTask(ctx context.Context, req *api.ModifyTaskRequest) (*api.TaskDetails, error) {
	auth_err := srv.authenticator.CheckAuthentication(ctx, FLEA_TASK_GEN)
	if auth_err != nil {
		return nil, auth_err
	}
	log.Println("ModifyTask:", req)
	err := srv.storage.ModifyTask(ctx, *req)
	if err != nil {
		return nil, err
	}
	// Can't call srv.GetTask because it authenticates against a diff token.
	resp, err := srv.storage.GetTask(ctx, req.TaskId)
	return &resp, err
}

func (srv *flea_server) ListTasks(ctx context.Context, req *api.ListTasksRequest) (*api.ListTasksResponse, error) {
	auth_err := srv.authenticator.CheckAuthentication(ctx, FLEA_CLIENT)
	if auth_err != nil {
		return nil, auth_err
	}
	log.Println("List tasks:", req)
	if req.CheckpointId != "" && req.HyperparametersId == "" {
		return nil, storage.ErrInvalidModelHyperparamsCheckpointCombo
	}
	if req.HyperparametersId != "" && req.ModelId == "" {
		return nil, storage.ErrInvalidModelHyperparamsCheckpointCombo
	}
	resp, err := srv.storage.ListTasks(ctx, *req)
	return &resp, err
}

func (srv *flea_server) GetTask(ctx context.Context, req *api.GetTaskRequest) (*api.TaskDetails, error) {
	auth_err := srv.authenticator.CheckAuthentication(ctx, FLEA_CLIENT)
	if auth_err != nil {
		return nil, auth_err
	}
	resp, err := srv.storage.GetTask(ctx, req.TaskId)
	return &resp, err
}

func (srv *flea_server) StartTask(ctx context.Context, req *api.StartTaskRequest) (*api.StartTaskResponse, error) {
	auth_err := srv.authenticator.CheckAuthentication(ctx, FLEA_CLIENT)
	if auth_err != nil {
		return nil, auth_err
	}
	resp, err := srv.storage.StartTask(ctx, req.TaskId)
	return &resp, err
}
