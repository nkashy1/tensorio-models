package main

import (
	"context"
	"github.com/doc-ai/tensorio-models/api"
	"log"
)

type server struct {
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

func (srv *server) CreateModel(context.Context, *api.CreateModelRequest) (*api.CreateModelResponse, error) {
	resp := &api.CreateModelResponse{
		ResourcePath: "/models/Manna",
	}

	return resp, nil
}

func (srv *server) GetModels(ctx context.Context, msg *api.GetModelsRequest) (*api.GetModelsResponse, error) {
	log.Println("Responding to: ", msg)
	resp := &api.GetModelsResponse{
		ModelIds: []string{"HappyFace", "PhenomenalFace"},
	}
	return resp, nil
}
