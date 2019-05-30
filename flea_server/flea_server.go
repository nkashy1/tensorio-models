package flea_server

import (
	"context"
	"errors"
	"net"
	"net/http"

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

// NewServer - Creates an api.FleaServer which handles gRPC requests using a given
// storage.FleaStorage backend
func NewServer(storage storage.FleaStorage, authenticator authentication.Authenticator) api.FleaServer {
	return &flea_server{
		storage:       storage,
		authenticator: authenticator,
	}
}

func startGrpcServer(apiServer api.FleaServer, serverAddress string, authInterceptor grpc.UnaryServerInterceptor) {
	log.Println("Starting FLEA gRPC on:", serverAddress)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(authInterceptor))
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

const (
	FleaAdmin   authentication.AuthenticationTokenType = "FleaAdmin"
	FleaClient  authentication.AuthenticationTokenType = "FleaClient"
	FleaTaskGen authentication.AuthenticationTokenType = "FleaTaskGen"
)

// CreateMethodToTokenTypeMap - must include a token type or NoAuthentication for all RPC methods.
func CreateMethodToTokenTypeMap() authentication.MethodToAuthenticationTokenType {
	return authentication.MethodToAuthenticationTokenType{
		"/api.Flea/Healthz": authentication.NoAuthentication,
		"/api.Flea/Config":  authentication.NoAuthentication,

		"/api.Flea/CreateTask": FleaTaskGen,
		"/api.Flea/ModifyTask": FleaTaskGen,

		"/api.Flea/Admin": FleaAdmin,

		"/api.Flea/GetTask":   FleaClient,
		"/api.Flea/ListTasks": FleaClient,
		"/api.Flea/StartTask": FleaClient,
		"/api.Flea/JobError":  FleaClient,
	}
}

// StartGrpcAndProxyServer - Given a repository storage backend, this function starts a
// new gRPC and JSON-RPC server in separate threads and waits until a message is received on the stopRequested channel.
func StartGrpcAndProxyServer(storage storage.FleaStorage,
	grpcServerAddress string, jsonServerAddress string,
	authenticator authentication.Authenticator,
	stopRequested <-chan string) {
	apiServer := NewServer(storage, authenticator)
	authInterceptor := authentication.CreateGRPCInterceptor(authenticator,
		CreateMethodToTokenTypeMap(),
	)
	go startGrpcServer(apiServer, grpcServerAddress, authInterceptor)
	go startProxyServer(grpcServerAddress, jsonServerAddress)
	stopReason := <-stopRequested
	log.Println("Stopping server due to:", stopReason)
}

func (srv *flea_server) Healthz(ctx context.Context, req *api.HealthCheckRequest) (*api.HealthCheckResponse, error) {
	resp := &api.HealthCheckResponse{
		Status: api.HealthCheckResponse_SERVING,
	}
	return resp, nil
}

func (srv *flea_server) Config(ctx context.Context, req *api.ConfigRequest) (*api.ConfigResponse, error) {
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

func (srv *flea_server) Admin(ctx context.Context, req *api.AdminRequest) (*api.GenericResponse, error) {
	if req.Type != api.AdminRequest_RELOAD_TOKENS {
		return nil, errors.New("Unknown admin request type!")
	}
	err := srv.authenticator.ReloadAuthenticationTokens(ctx)
	if err != nil {
		return nil, err
	}
	return &api.GenericResponse{Message: "Updated Authentication Tokens"}, nil
}

func (srv *flea_server) JobError(ctx context.Context, req *api.JobErrorRequest) (*api.GenericResponse, error) {
	if req.ErrorMessage == "" {
		err := errors.New("Expected non-empty errorMessage")
		return nil, err
	}
	err := srv.storage.AddJobError(ctx, *req)
	if err != nil {
		return nil, err
	}
	return &api.GenericResponse{Message: "Thank you for the error report."}, nil
}

func (srv *flea_server) CreateTask(ctx context.Context, req *api.TaskDetails) (*api.TaskDetails, error) {
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
	err := srv.storage.ModifyTask(ctx, *req)
	if err != nil {
		return nil, err
	}
	// Can't call srv.GetTask because it authenticates against a diff token.
	resp, err := srv.storage.GetTask(ctx, req.TaskId)
	return &resp, err
}

func (srv *flea_server) ListTasks(ctx context.Context, req *api.ListTasksRequest) (*api.ListTasksResponse, error) {
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
	resp, err := srv.storage.GetTask(ctx, req.TaskId)
	return &resp, err
}

func (srv *flea_server) StartTask(ctx context.Context, req *api.StartTaskRequest) (*api.StartTaskResponse, error) {
	resp, err := srv.storage.StartTask(ctx, req.TaskId)
	return &resp, err
}
