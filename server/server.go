package server

import (
	"context"
	"fmt"

	"github.com/doc-ai/tensorio-models/api"
	"github.com/doc-ai/tensorio-models/storage"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	storage storage.RepositoryStorage
}

func NewServer(storage storage.RepositoryStorage) api.RepositoryServer {
	return &server{
		storage: storage,
	}
}

func (srv *server) ListModels(ctx context.Context, req *api.ListModelsRequest) (*api.ListModelsResponse, error) {
	log.Println(req.Marker)
	res := &api.ListModelsResponse{
		ModelIds: []string{"HappyFace", "PhenomenalFace"},
	}
	return res, nil
}

func (srv *server) GetModel(ctx context.Context, req *api.GetModelRequest) (*api.GetModelResponse, error) {
	resp := &api.GetModelResponse{
		Model: &api.Model{
			ModelId:                  req.ModelId,
			Description:              "Accepts images of food and returns a number between 0 and 1 signifying how healthy a person off the street would consider that food to be (0 being unhealthy and 1 being healthy).",
			CanonicalHyperParameters: "batch-8-et-v2-140-224-ing-rate-1e-5",
		},
	}
	return resp, nil
}

func (srv *server) CreateModel(ctx context.Context, req *api.CreateModelRequest) (*api.CreateModelResponse, error) {
	resp := &api.CreateModelResponse{
		ResourcePath: fmt.Sprintf("/models/%s", req.Model.ModelId),
	}

	return resp, nil
}

func (srv *server) UpdateModel(ctx context.Context, req *api.UpdateModelRequest) (*api.UpdateModelResponse, error) {
	log.Println(ctx, req)
	if req.ModelId != "Manna" {
		return nil, status.Error(codes.NotFound, "Resource not found")
	}
	resp := &api.UpdateModelResponse{
		Model: &api.Model{
			ModelId:                  req.ModelId,
			Description:              "Accepts images of food and returns a number between 0 and 1 signifying how healthy a person off the street would consider that food to be (0 being unhealthy and 1 being healthy).",
			CanonicalHyperParameters: "batch-9-ion-resnet-v2-ing-rate-1e-7",
		},
	}
	return resp, nil
}

func (srv *server) ListHyperParameters(ctx context.Context, req *api.ListHyperParametersRequest) (*api.ListHyperParametersResponse, error) {
	log.Println(req)

	resp := &api.ListHyperParametersResponse{
		HyperParametersIds: []string{"batch-9-2-0-1-5", "batch-9-2-0-1-0", "batch-9-2-0-0-5"},
	}
	return resp, nil
}

func (srv *server) CreateHyperParameters(ctx context.Context, req *api.CreateHyperParametersRequest) (*api.CreateHyperParametersResponse, error) {
	log.Println(req)
	resp := &api.CreateHyperParametersResponse{
		ResourcePath: fmt.Sprintf("/models/%s/hyperparameters/%s", req.ModelId, req.HyperParameterId),
	}
	return resp, nil
}

func (srv *server) GetHyperParameters(ctx context.Context, req *api.GetHyperParametersRequest) (*api.GetHyperParametersResponse, error) {
	log.Println(req)
	resp := &api.GetHyperParametersResponse{
		ModelId:             req.ModelId,
		HyperParametersId:   req.HyperParametersId,
		UpgradeTo:           "",
		CanonicalCheckpoint: "model.ckpt-321312",
		HyperParameters: map[string]string{
			"architecture":                  "inception-resnet-v3",
			"batch":                         "9",
			"training-set-entropy-cutoff":   "2.0",
			"evaluation-set-entropy-cutoff": "2.0",
		},
	}
	return resp, nil
}

func (srv *server) UpdateHyperParameters(ctx context.Context, req *api.UpdateHyperParametersRequest) (*api.UpdateHyperParametersResponse, error) {
	log.Println(req)
	resp := &api.UpdateHyperParametersResponse{
		ModelId:             req.ModelId,
		HyperParametersId:   req.HyperParametersId,
		UpgradeTo:           req.UpgradeTo,
		CanonicalCheckpoint: "model.ckpt-321312",
		HyperParameters: map[string]string{
			"architecture":                  "inception-resnet-v3",
			"batch":                         "9",
			"training-set-entropy-cutoff":   "2.0",
			"evaluation-set-entropy-cutoff": "2.0",
		},
	}
	return resp, nil
}

func (srv *server) ListCheckpoints(ctx context.Context, req *api.ListCheckpointsRequest) (*api.ListCheckpointsResponse, error) {
	log.Println(ctx)
	resp := &api.ListCheckpointsResponse{
		CheckpointIds: []string{"model.ckpt-321312", "model.ckpt-320210", "model.ckpt-319117"},
	}
	return resp, nil
}

func (srv *server) CreateCheckpoint(ctx context.Context, req *api.CreateCheckpointRequest) (*api.CreateCheckpointResponse, error) {
	log.Println(req)
	resourcePath := getCheckpointResourcePath(req.ModelId, req.HyperParametersId, req.CheckpointId)
	resp := &api.CreateCheckpointResponse{
		ResourcePath: resourcePath,
	}
	return resp, nil
}

func (srv *server) GetCheckpoint(ctx context.Context, req *api.GetCheckpointRequest) (*api.GetCheckpointResponse, error) {
	resourcePath := getCheckpointResourcePath(req.ModelId, req.HyperParametersId, req.CheckpointId)
	resp := &api.GetCheckpointResponse{
		ResourcePath: resourcePath,
		Link:         "https://storage.googleapis.com/doc-ai-models/happy-face/batch-9-2-0-9-2-0/model.ckpt-322405.zip",
		Info:         map[string]string{"standard-1-accuracy": "0.934"},
	}
	return resp, nil
}

func getCheckpointResourcePath(modelId, hyperParametersId, checkpointId string) string {
	resourcePath := fmt.Sprintf("models/%s/hyperparameters/%s/checkpoints/%s", modelId, hyperParametersId, checkpointId)
	return resourcePath
}
