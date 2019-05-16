package flea_server

import (
	"context"
	"net"

	"github.com/doc-ai/tensorio-models/api"
	"github.com/doc-ai/tensorio-models/server"
	"github.com/doc-ai/tensorio-models/storage"
	"github.com/doc-ai/tensorio-models/storage/gcs"
	"github.com/doc-ai/tensorio-models/storage/memory"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type flea_server struct {
	storage storage.FleaStorage
}

// NewServer - Creates an api.RepositoryServer which handles gRPC requests using a given
// storage.RepositoryStorage backend
func NewServer(storage storage.FleaStorage) api.FleaServer {
	return &flea_server{storage: storage}
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

// StartGrpcAndProxyServer - Given a repository storage backend, this function starts a
// new gRPC and JSON-RPC server in separate threads and waits until a message is received on the stopRequested channel.
func StartGrpcAndProxyServer(storage storage.FleaStorage,
	grpcServerAddress string, jsonServerAddress string,
	stopRequested <-chan string) {
	apiServer := NewServer(storage)
	go startGrpcServer(apiServer, grpcServerAddress)
	go server.StartProxyServer(grpcServerAddress, jsonServerAddress)
	stopReason := <-stopRequested
	log.Println("Stopping server due to:", stopReason)
}

func (srv *flea_server) Healthz(ctx context.Context, req *api.HealthCheckRequest) (*api.HealthCheckResponse, error) {
	log.Println(req)
	resp := &api.HealthCheckResponse{
		Status: api.HealthCheckResponse_SERVING,
	}
	return resp, nil
}

func (srv *flea_server) Config(ctx context.Context, req *api.ConfigRequest) (*api.ConfigResponse, error) {
	log.Println(req)
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

func (srv *flea_server) CreateTask(ctx context.Context, req *api.TaskDetails) (*api.TaskDetails, error) {
	if req.ModelId == "" {
		return nil, storage.ErrMissingModelId
	}
	if req.HyperparametersId == "" {
		return nil, storage.ErrMissingHyperparametersId
	}
	if req.CheckpointId == "" {
		return nil, storage.ErrMissingCheckpointId
	}
	if req.TaskId == "" {
		return nil, storage.ErrMissingTaskId
	}
	err := srv.storage.AddTask(ctx, *req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (srv *flea_server) ModifyTask(ctx context.Context, req *api.ModifyTaskRequest) (*api.TaskDetails, error) {
	err := srv.storage.ModifyTask(ctx, *req)
	if err != nil {
		return nil, err
	}
	return srv.GetTask(ctx, &api.GetTaskRequest{TaskId: req.TaskId})
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