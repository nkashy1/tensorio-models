package server_test

import (
	"context"
	"fmt"
	"github.com/doc-ai/tensorio-models/api"
	"github.com/doc-ai/tensorio-models/server"
	"github.com/doc-ai/tensorio-models/storage/memory"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func testingServer() api.RepositoryServer {
	storage := memory.NewMemoryRepositoryStorage()
	server := server.NewServer(storage)
	return server
}

// Tests that models can be successfully created and listed against a fresh memory RepositoryStorage
// backend.
// Also tests that attempts to create duplicated models result in errors.
func TestCreateModel(t *testing.T) {
	server := testingServer()

	modelRequests := make([]api.CreateModelRequest, 5)
	for i := range modelRequests {
		modelID := fmt.Sprintf("test-model-%d", i)
		description := fmt.Sprintf("This is test model %d", i)
		model := api.CreateModelRequest{
			Model: &api.Model{
				ModelId:     modelID,
				Description: description,
			},
		}
		modelRequests[i] = model
	}

	ctx := context.Background()

	// TODO: Instead of doing ListModels on the server, we should inspect server.storage directly
	listModelsRequest := api.ListModelsRequest{
		MaxItems: 20,
	}
	for i, req := range modelRequests {
		listModelsResponse, err := server.ListModels(ctx, &listModelsRequest)
		models := listModelsResponse.ModelIds
		if err != nil {
			t.Error(err)
		}
		if len(models) != i {
			t.Errorf("Incorrect number of models in storage; expected: %d, actual: %d", i, len(models))
		}
		createModelResponse, err := server.CreateModel(ctx, &req)
		if err != nil {
			t.Error(err)
		}
		expectedResourcePath := fmt.Sprintf("/models/%s", req.Model.ModelId)
		if createModelResponse.ResourcePath != expectedResourcePath {
			t.Errorf("Incorrect resource path for created model; expected: %s, actual: %s", expectedResourcePath, createModelResponse.ResourcePath)
		}
	}

	// Creation with a duplicated request should fail
	_, err := server.CreateModel(ctx, &modelRequests[0])
	if err == nil {
		t.Error("Server did not error out on creation of duplicate model")
	}
}

// Tests that models are correctly listed (pagination behaviour)
func TestListModels(t *testing.T) {
	server := testingServer()

	modelRequests := make([]api.CreateModelRequest, 21)
	for i := range modelRequests {
		modelID := fmt.Sprintf("test-model-%d", i)
		description := fmt.Sprintf("This is test model %d", i)
		model := api.CreateModelRequest{
			Model: &api.Model{
				ModelId:     modelID,
				Description: description,
			},
		}
		modelRequests[i] = model
	}
	ctx := context.Background()
	modelIDs := make([]string, len(modelRequests))
	for i, req := range modelRequests {
		modelIDs[i] = req.Model.ModelId
		_, err := server.CreateModel(ctx, &req)
		if err != nil {
			t.Error(err)
		}
	}
	// NOTE: ModelIDs are sorted lexicographically, not chronologically!
	sort.Strings(modelIDs)

	type ListModelsTest struct {
		Server           *api.RepositoryServer
		Marker           string
		MaxItems         int32
		ExpectedModelIds []string
	}

	tests := []ListModelsTest{
		{
			Server:           &server,
			MaxItems:         int32(5),
			ExpectedModelIds: modelIDs[0:5],
		},
		{
			Server:           &server,
			Marker:           modelIDs[2],
			MaxItems:         int32(5),
			ExpectedModelIds: modelIDs[2:7],
		},
		{
			Server:           &server,
			Marker:           modelIDs[16],
			MaxItems:         int32(5),
			ExpectedModelIds: modelIDs[16:21],
		},
		{
			Server:           &server,
			Marker:           modelIDs[16],
			MaxItems:         int32(6),
			ExpectedModelIds: modelIDs[16:21],
		},
		// TODO(frederick): Specification says that list endpoints should return items AFTER marker,
		// not after and including marker. No need to change behaviour, just make the two consistent.
		{
			Server:           &server,
			Marker:           modelIDs[0],
			MaxItems:         int32(20),
			ExpectedModelIds: modelIDs[0:20],
		},
	}

	for i, test := range tests {
		listModelsRequest := api.ListModelsRequest{
			Marker:   test.Marker,
			MaxItems: test.MaxItems,
		}

		tsrv := *test.Server
		listModelsResponse, err := tsrv.ListModels(ctx, &listModelsRequest)
		if err != nil {
			t.Error(err)
		}
		assert.Equalf(t, test.ExpectedModelIds, listModelsResponse.ModelIds, "TestListModels %d: ListModels request returned incorrect ModelIds", i)
	}
}

// Tests that model update behaviour is correct
func TestUpdateModel(t *testing.T) {
	server := testingServer()

	model := api.CreateModelRequest{
		Model: &api.Model{
			ModelId:     "test-model",
			Description: "This is a test",
		},
	}

	ctx := context.Background()

	_, err := server.CreateModel(ctx, &model)
	if err != nil {
		t.Error(err)
	}

	updateModelRequest := api.UpdateModelRequest{
		ModelId: model.Model.ModelId,
		Model: &api.Model{
			ModelId:     "test-model",
			Description: "This is only a test",
		},
	}
	updateModelResponse, err := server.UpdateModel(ctx, &updateModelRequest)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, updateModelRequest.Model, updateModelResponse.Model, "UpdateModel models in request and response do not agree")
}

// Creates a model and tests that GetModel returns the expected information
func TestGetModel(t *testing.T) {
	server := testingServer()

	model := api.CreateModelRequest{
		Model: &api.Model{
			ModelId:     "test-model",
			Description: "This is a test",
		},
	}
	ctx := context.Background()
	_, err := server.CreateModel(ctx, &model)
	if err != nil {
		t.Error(err)
	}

	getModelRequest := api.GetModelRequest{ModelId: model.Model.ModelId}
	getModelResponse, err := server.GetModel(ctx, &getModelRequest)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, model.Model, getModelResponse.Model, "Did not receive the expected model in GetModel response")
}
