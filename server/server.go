package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/doc-ai/tensorio-models/api"
	"github.com/doc-ai/tensorio-models/storage"
	"github.com/doc-ai/tensorio-models/storage/gcs"
	"github.com/doc-ai/tensorio-models/storage/memory"
	"github.com/golang/protobuf/ptypes"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	storage storage.RepositoryStorage
}

// NewServer - Creates an api.RepositoryServer which handles gRPC requests using a given
// storage.RepositoryStorage backend
func NewServer(storage storage.RepositoryStorage) api.RepositoryServer {
	return &server{storage: storage}
}

// IsValidIDChar - returns whether c is a valid character - alphanumeric, _ and -
func IsValidIDChar(c byte) bool {
	if (c >= 'A') && (c <= 'Z') {
		return true
	}
	if (c >= 'a') && (c <= 'z') {
		return true
	}
	if (c >= '0') && (c <= '9') {
		return true
	}
	if (c == '-') || (c == '_') {
		return true
	}
	return false
}

// IsValidID - returns whether s is non-empty string of valid characters
func IsValidID(s string) bool {
	l := len(s)
	for i := 0; i < l; i++ {
		if !IsValidIDChar(s[i]) {
			return false
		}
	}
	return s != ""
}

func startGrpcServer(apiServer api.RepositoryServer, serverAddress string) {
	log.Println("Starting grpc on:", serverAddress)

	grpcServer := grpc.NewServer()
	lis, err := net.Listen("tcp", serverAddress)
	if err != nil {
		log.Fatalln(err)
	}

	api.RegisterRepositoryServer(grpcServer, apiServer)

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
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := api.RegisterRepositoryHandlerFromEndpoint(ctx, mux, grpcServerAddress, opts)
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
func StartGrpcAndProxyServer(storage storage.RepositoryStorage,
	grpcServerAddress string, jsonServerAddress string,
	stopRequested <-chan string) {
	apiServer := NewServer(storage)
	go startGrpcServer(apiServer, grpcServerAddress)
	go startProxyServer(grpcServerAddress, jsonServerAddress)
	stopReason := <-stopRequested
	log.Println("Stopping server due to:", stopReason)
}

func (srv *server) Healthz(ctx context.Context, req *api.HealthCheckRequest) (*api.HealthCheckResponse, error) {
	log.Println(req)
	resp := &api.HealthCheckResponse{
		Status: api.HealthCheckResponse_SERVING,
	}
	return resp, nil
}

func (srv *server) Config(ctx context.Context, req *api.ConfigRequest) (*api.ConfigResponse, error) {
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

func (srv *server) ListModels(ctx context.Context, req *api.ListModelsRequest) (*api.ListModelsResponse, error) {
	marker := req.Marker
	maxItems := int(req.MaxItems)
	if maxItems <= 0 {
		maxItems = 10
	}
	log.Printf("ListModels request - Marker: %s, MaxItems: %d", marker, maxItems)
	models, err := srv.storage.ListModels(ctx, marker, maxItems)
	if err != nil {
		log.Printf("ERROR: %v", err)
		grpcErr := status.Error(codes.Unavailable, "Could not retrieve models from storage")
		return nil, grpcErr
	}
	res := &api.ListModelsResponse{
		ModelIds: models,
	}
	return res, nil
}

func (srv *server) GetModel(ctx context.Context, req *api.GetModelRequest) (*api.GetModelResponse, error) {
	modelID := req.ModelId
	log.Printf("GetModel request - ModelId: %s", modelID)
	model, err := srv.storage.GetModel(ctx, modelID)
	if err != nil {
		log.Printf("ERROR: %v", err)
		message := fmt.Sprintf("Could not retrieve model (%s) from storage", modelID)
		grpcErr := status.Error(codes.Unavailable, message)
		return nil, grpcErr
	}
	respModel := api.Model{
		ModelId:                  model.ModelId,
		Details:                  model.Details,
		CanonicalHyperparameters: model.CanonicalHyperparameters,
	}
	resp := &api.GetModelResponse{
		Model: &respModel,
	}
	return resp, nil
}

func (srv *server) CreateModel(ctx context.Context, req *api.CreateModelRequest) (*api.CreateModelResponse, error) {
	model := req.Model
	log.Printf("CreateModel request - Model: %v", model)
	// Check that ModelId is a non-empty string
	if !IsValidID(model.ModelId) {
		grpcErr := status.Error(codes.InvalidArgument, "ModelId was invalid")
		return nil, grpcErr
	}
	// Check that Details field in request model is not nil
	if model.Details == "" {
		grpcErr := status.Error(codes.InvalidArgument, "Details should be specified in CreateModel request")
		return nil, grpcErr
	}
	storageModel := storage.Model{
		ModelId:                  model.ModelId,
		Details:                  model.Details,
		CanonicalHyperparameters: model.CanonicalHyperparameters,
	}
	err := srv.storage.AddModel(ctx, storageModel)
	if err != nil {
		log.Printf("ERROR: %v", err)
		grpcErr := status.Error(codes.Unavailable, "Could not store model")
		return nil, grpcErr
	}
	resourcePath := fmt.Sprintf("/models/%s", storageModel.ModelId)
	resp := &api.CreateModelResponse{ResourcePath: resourcePath}
	return resp, nil
}

func (srv *server) UpdateModel(ctx context.Context, req *api.UpdateModelRequest) (*api.UpdateModelResponse, error) {
	modelID := req.ModelId
	model := req.Model
	if modelID == "" {
		return nil, api.MissingRequiredFieldError("modelId", "model id to update").Err()
	}
	if model == nil {
		return nil, api.MissingRequiredFieldError("model", "model to update").Err()
	}
	if req.Model.ModelId != "" {
		if req.ModelId != req.Model.ModelId {
			return nil, api.InvalidFieldValueError("request.model.modelId", "request.modelId != request.model.modelId").Err()
		}
	}
	log.Printf("UpdateModel request - ModelId: %s, Model: %v", modelID, model)
	storedModel, err := srv.storage.GetModel(ctx, modelID)
	if err != nil {
		log.Printf("ERROR: %v", err)
		message := fmt.Sprintf("Could not retrieve model (%s) from storage", modelID)
		grpcErr := status.Error(codes.Unavailable, message)
		return nil, grpcErr
	}
	updatedModel := storedModel
	if model.Details != "" {
		updatedModel.Details = model.Details
	}
	if model.CanonicalHyperparameters != "" {
		updatedModel.CanonicalHyperparameters = model.CanonicalHyperparameters
	}
	newlyStoredModel, err := srv.storage.UpdateModel(ctx, updatedModel)
	if err != nil {
		log.Printf("ERROR: %v", err)
		message := fmt.Sprintf("Could not update model (%s) in storage", modelID)
		grpcErr := status.Error(codes.Unavailable, message)
		return nil, grpcErr
	}
	resp := &api.UpdateModelResponse{
		Model: &api.Model{
			ModelId:                  newlyStoredModel.ModelId,
			Details:                  newlyStoredModel.Details,
			CanonicalHyperparameters: newlyStoredModel.CanonicalHyperparameters,
		},
	}
	return resp, nil
}

func (srv *server) ListHyperparameters(ctx context.Context, req *api.ListHyperparametersRequest) (*api.ListHyperparametersResponse, error) {
	modelID := req.ModelId
	marker := req.Marker
	maxItems := int(req.MaxItems)
	if maxItems <= 0 {
		maxItems = 10
	}
	log.Printf("ListHyperparameters request - ModelId: %s, Marker: %s, MaxItems: %d", modelID, marker, maxItems)
	hyperparametersStoragePaths, err := srv.storage.ListHyperparameters(ctx, modelID, marker, maxItems)
	if err != nil {
		log.Printf("ERROR: %v", err)
		message := fmt.Sprintf("Could not list hyperparameters for model (%s) in storage", modelID)
		grpcErr := status.Error(codes.Unavailable, message)
		return nil, grpcErr
	}
	hyperparametersIDs := make([]string, len(hyperparametersStoragePaths))
	for i, path := range hyperparametersStoragePaths {
		hyperparametersIDs[i] = getTerminalResourceFromStoragePath(path)
	}
	resp := &api.ListHyperparametersResponse{
		ModelId:            modelID,
		HyperparametersIds: hyperparametersIDs,
	}
	return resp, nil
}

func (srv *server) CreateHyperparameters(ctx context.Context, req *api.CreateHyperparametersRequest) (*api.CreateHyperparametersResponse, error) {
	modelID := req.ModelId
	hyperparametersID := req.HyperparametersId
	if !IsValidID(hyperparametersID) {
		grpcErr := status.Error(codes.InvalidArgument, "hyperparametersID is invalid")
		return nil, grpcErr
	}
	canonicalCheckpoint := req.CanonicalCheckpoint
	hyperparameters := req.Hyperparameters
	log.Printf("CreateHyperparameters request - ModelId: %s, HyperparametersId: %s, CanonicalCheckpoint: %s, Hyperparameters: %v", modelID, hyperparametersID, canonicalCheckpoint, hyperparameters)
	storageHyperparameters := storage.Hyperparameters{
		ModelId:             modelID,
		HyperparametersId:   hyperparametersID,
		CanonicalCheckpoint: canonicalCheckpoint,
		Hyperparameters:     hyperparameters,
	}
	err := srv.storage.AddHyperparameters(ctx, storageHyperparameters)
	if err != nil {
		log.Printf("ERROR: %v", err)
		message := fmt.Sprintf("Could not store hyperparameters (%v) in storage", storageHyperparameters)
		grpcErr := status.Error(codes.Unavailable, message)
		return nil, grpcErr
	}
	resourcePath := fmt.Sprintf("/models/%s/hyperparameters/%s", modelID, hyperparametersID)
	resp := &api.CreateHyperparametersResponse{
		ResourcePath: resourcePath,
	}
	return resp, nil
}

func (srv *server) GetHyperparameters(ctx context.Context, req *api.GetHyperparametersRequest) (*api.GetHyperparametersResponse, error) {
	modelID := req.ModelId
	hyperparametersID := req.HyperparametersId
	log.Printf("GetHyperparameters request - ModelId: %s, HyperparametersId: %s", modelID, hyperparametersID)
	storedHyperparameters, err := srv.storage.GetHyperparameters(ctx, modelID, hyperparametersID)
	if err != nil {
		log.Printf("ERROR: %v", err)
		message := fmt.Sprintf("Could not get hyperparameters (%s) for model (%s) from storage", hyperparametersID, modelID)
		grpcErr := status.Error(codes.Unavailable, message)
		return nil, grpcErr
	}
	resp := &api.GetHyperparametersResponse{
		ModelId:             storedHyperparameters.ModelId,
		HyperparametersId:   storedHyperparameters.HyperparametersId,
		UpgradeTo:           storedHyperparameters.UpgradeTo,
		CanonicalCheckpoint: storedHyperparameters.CanonicalCheckpoint,
		Hyperparameters:     storedHyperparameters.Hyperparameters,
	}
	return resp, nil
}

func (srv *server) UpdateHyperparameters(ctx context.Context, req *api.UpdateHyperparametersRequest) (*api.UpdateHyperparametersResponse, error) {
	modelID := req.ModelId
	hyperparametersID := req.HyperparametersId
	upgradeTo := req.UpgradeTo
	canonicalCheckpoint := req.CanonicalCheckpoint
	hyperparameters := req.Hyperparameters
	log.Printf("UpdateHyperparameters request - ModelId: %s, HyperparametersId: %s, CanonicalCheckpoint: %s, Hyperparameters: %v", modelID, hyperparametersID, canonicalCheckpoint, hyperparameters)

	existingHyperparameters, err := srv.storage.GetHyperparameters(ctx, modelID, hyperparametersID)
	if err != nil {
		log.Printf("ERROR: %v", err)
		message := fmt.Sprintf("Could not get hyperparameters (%s) for model (%s) from storage", hyperparametersID, modelID)
		grpcErr := status.Error(codes.Unavailable, message)
		return nil, grpcErr
	}

	updatedHyperparameters := existingHyperparameters
	if canonicalCheckpoint != "" {
		updatedHyperparameters.CanonicalCheckpoint = canonicalCheckpoint
	}
	if upgradeTo != "" {
		updatedHyperparameters.UpgradeTo = upgradeTo
	}
	for k, v := range hyperparameters {
		updatedHyperparameters.Hyperparameters[k] = v
	}
	storedHyperparameters, err := srv.storage.UpdateHyperparameters(ctx, updatedHyperparameters)
	if err != nil {
		log.Printf("ERROR: %v", err)
		message := fmt.Sprintf("Could not store hyperparameters (%v) in storage", updatedHyperparameters)
		grpcErr := status.Error(codes.Unavailable, message)
		return nil, grpcErr
	}

	resp := &api.UpdateHyperparametersResponse{
		ModelId:             storedHyperparameters.ModelId,
		HyperparametersId:   storedHyperparameters.HyperparametersId,
		UpgradeTo:           "",
		CanonicalCheckpoint: storedHyperparameters.CanonicalCheckpoint,
		Hyperparameters:     storedHyperparameters.Hyperparameters,
	}
	return resp, nil
}

func (srv *server) ListCheckpoints(ctx context.Context, req *api.ListCheckpointsRequest) (*api.ListCheckpointsResponse, error) {
	modelID := req.ModelId
	hyperparametersID := req.HyperparametersId
	marker := req.Marker
	maxItems := int(req.MaxItems)
	if maxItems <= 0 {
		maxItems = 10
	}
	log.Printf("ListCheckpoints request - ModelId: %s, HyperparametersId: %s, Marker: %s, MaxItems: %d", modelID, hyperparametersID, marker, maxItems)
	checkpointStoragePaths, err := srv.storage.ListCheckpoints(ctx, modelID, hyperparametersID, marker, maxItems)
	if err != nil {
		log.Printf("ERROR: %v", err)
		message := fmt.Sprintf("Could not list checkpoints for model (%s) and hyperparameters (%s) in storage", modelID, hyperparametersID)
		grpcErr := status.Error(codes.Unavailable, message)
		return nil, grpcErr
	}
	checkpointIDs := make([]string, len(checkpointStoragePaths))
	for i, path := range checkpointStoragePaths {
		checkpointIDs[i] = getTerminalResourceFromStoragePath(path)
	}
	resp := &api.ListCheckpointsResponse{
		ModelId:           modelID,
		HyperparametersId: hyperparametersID,
		CheckpointIds:     checkpointIDs,
	}
	return resp, nil
}

func (srv *server) CreateCheckpoint(ctx context.Context, req *api.CreateCheckpointRequest) (*api.CreateCheckpointResponse, error) {
	modelID := req.ModelId
	hyperparametersID := req.HyperparametersId
	checkpointID := req.CheckpointId
	if !IsValidID(checkpointID) {
		grpcErr := status.Error(codes.InvalidArgument, "checkpointId is invalid")
		return nil, grpcErr
	}
	link := req.Link
	log.Printf("CreateCheckpoint request - ModelId: %s, HyperparametersId: %s, CheckpointId: %s, Link: %s", modelID, hyperparametersID, checkpointID, link)
	storageCheckpoint := storage.Checkpoint{
		ModelId:           modelID,
		HyperparametersId: hyperparametersID,
		CheckpointId:      checkpointID,
		CreatedAt:         time.Now(),
		Link:              link,
		Info:              req.Info,
	}
	err := srv.storage.AddCheckpoint(ctx, storageCheckpoint)
	if err != nil {
		log.Printf("ERROR: %v", err)
		message := fmt.Sprintf("Could not store checkpoint (%v) in storage", storageCheckpoint)
		grpcErr := status.Error(codes.Unavailable, message)
		return nil, grpcErr
	}
	resourcePath := getCheckpointResourcePath(modelID, hyperparametersID, checkpointID)
	resp := &api.CreateCheckpointResponse{
		ResourcePath: resourcePath,
	}
	return resp, nil
}

func (srv *server) GetCheckpoint(ctx context.Context, req *api.GetCheckpointRequest) (*api.GetCheckpointResponse, error) {
	modelID := req.ModelId
	hyperparametersID := req.HyperparametersId
	checkpointID := req.CheckpointId
	log.Printf("GetCheckpoint request - ModelId: %s, HyperparametersId: %s, CheckpointId: %s", modelID, hyperparametersID, checkpointID)
	storedCheckpoint, err := srv.storage.GetCheckpoint(ctx, modelID, hyperparametersID, checkpointID)
	if err != nil {
		log.Printf("ERROR: %v", err)
		message := fmt.Sprintf("Could not get checkpoint (%s) of hyperparameters (%s) for model (%s) from storage", checkpointID, hyperparametersID, modelID)
		grpcErr := status.Error(codes.Unavailable, message)
		return nil, grpcErr
	}
	resourcePath := getCheckpointResourcePath(modelID, hyperparametersID, checkpointID)
	createdAt, err := ptypes.TimestampProto(storedCheckpoint.CreatedAt)
	if err != nil {
		log.Error("unable to serialize CreatedAt")
		return nil, err
	}
	resp := &api.GetCheckpointResponse{
		ResourcePath:      resourcePath,
		Link:              storedCheckpoint.Link,
		CreatedAt:         createdAt,
		Info:              storedCheckpoint.Info,
		ModelId:           modelID,
		HyperparametersId: hyperparametersID,
		CheckpointId:      checkpointID,
	}
	return resp, nil
}

func getCheckpointResourcePath(modelID, hyperparametersID, checkpointID string) string {
	resourcePath := fmt.Sprintf("/models/%s/hyperparameters/%s/checkpoints/%s", modelID, hyperparametersID, checkpointID)
	return resourcePath
}

// RepositoryStorage implementations return resources in the form:
// <modelId>, <modelId>:<hyperparametersId>, <modelId>:<hyperparametersId>:<checkpointId>
// This function takes input in those formats and returns (respectively):
// <modelId>, <hyperparametersId>, <checkpointId>
func getTerminalResourceFromStoragePath(storagePath string) string {
	storageDelimiter := ":"
	components := strings.Split(storagePath, storageDelimiter)
	terminalComponent := components[len(components)-1]
	return terminalComponent
}
