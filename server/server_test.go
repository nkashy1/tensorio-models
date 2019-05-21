package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/doc-ai/tensorio-models/api"
	"github.com/doc-ai/tensorio-models/common"
	"github.com/doc-ai/tensorio-models/server"

	"github.com/doc-ai/tensorio-models/storage/memory"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
)

func testingServer() api.RepositoryServer {
	storage := memory.NewMemoryRepositoryStorage()
	srv := server.NewServer(storage)
	return srv
}

// Tests that models can be successfully created and listed against a fresh memory RepositoryStorage
// backend.
// Also tests that attempts to create duplicated models result in errors.
func TestCreateModelAndListModels(t *testing.T) {
	srv := testingServer()
	modelRequests := make([]api.CreateModelRequest, 5)
	for i := range modelRequests {
		modelID := fmt.Sprintf("test-model-%d", i)
		details := fmt.Sprintf("This is test model %d", i)
		model := api.CreateModelRequest{
			Model: &api.Model{
				ModelId: modelID,
				Details: details,
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
		listModelsResponse, err := srv.ListModels(ctx, &listModelsRequest)
		assert.NoError(t, err)
		models := listModelsResponse.ModelIds
		if len(models) != i {
			t.Errorf("Incorrect number of models in storage; expected: %d, actual: %d", i, len(models))
		}
		createModelResponse, err := srv.CreateModel(ctx, &req)
		if err != nil {
			t.Error(err)
		}
		expectedResourcePath := fmt.Sprintf("/models/%s", req.Model.ModelId)
		if createModelResponse.ResourcePath != expectedResourcePath {
			t.Errorf("Incorrect resource path for created model; expected: %s, actual: %s", expectedResourcePath, createModelResponse.ResourcePath)
		}
	}

	// Creation with a duplicated request should fail
	_, err := srv.CreateModel(ctx, &modelRequests[0])
	if err == nil {
		t.Error("Server did not error out on creation of duplicate model")
	}
}

// Tests that models are correctly listed (pagination behaviour)
func TestListModels(t *testing.T) {
	srv := testingServer()

	modelRequests := make([]api.CreateModelRequest, 21)
	for i := range modelRequests {
		modelID := fmt.Sprintf("test-model-%d", i)
		details := fmt.Sprintf("This is test model %d", i)
		model := api.CreateModelRequest{
			Model: &api.Model{
				ModelId: modelID,
				Details: details,
			},
		}
		modelRequests[i] = model
	}
	ctx := context.Background()
	modelIDs := make([]string, len(modelRequests))
	for i, req := range modelRequests {
		modelIDs[i] = req.Model.ModelId
		_, err := srv.CreateModel(ctx, &req)
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
			Server:           &srv,
			MaxItems:         int32(5),
			ExpectedModelIds: modelIDs[0:5],
		},
		{
			Server:           &srv,
			Marker:           modelIDs[1],
			MaxItems:         int32(5),
			ExpectedModelIds: modelIDs[2:7],
		},
		{
			Server:           &srv,
			Marker:           modelIDs[15],
			MaxItems:         int32(5),
			ExpectedModelIds: modelIDs[16:21],
		},
		{
			Server:           &srv,
			Marker:           modelIDs[15],
			MaxItems:         int32(6),
			ExpectedModelIds: modelIDs[16:21],
		},
		// TODO(frederick): Specification says that list endpoints should return items AFTER marker,
		// not after and including marker. No need to change behaviour, just make the two consistent.
		{
			Server:           &srv,
			Marker:           modelIDs[0],
			MaxItems:         int32(20),
			ExpectedModelIds: modelIDs[1:21],
		},
		{
			Server:           &srv,
			Marker:           modelIDs[0],
			ExpectedModelIds: modelIDs[1:11],
		},
		{
			Server:           &srv,
			Marker:           modelIDs[0],
			MaxItems:         0,
			ExpectedModelIds: modelIDs[1:11],
		},
		{
			Server:           &srv,
			Marker:           modelIDs[0],
			MaxItems:         -10,
			ExpectedModelIds: modelIDs[1:11],
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

		if t.Failed() {
			break
		}
	}
}

// Tests that model update behaviour is correct
func TestUpdateModel(t *testing.T) {
	srv := testingServer()

	model := api.CreateModelRequest{
		Model: &api.Model{
			ModelId: "test-model",
			Details: "This is a test",
		},
	}

	ctx := context.Background()

	_, err := srv.CreateModel(ctx, &model)
	if err != nil {
		t.Error(err)
	}

	updateModelRequest := api.UpdateModelRequest{
		ModelId: model.Model.ModelId,
		Model: &api.Model{
			ModelId: "test-model",
			Details: "This is only a test",
		},
	}
	updateModelResponse, err := srv.UpdateModel(ctx, &updateModelRequest)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, updateModelRequest.Model, updateModelResponse.Model, "UpdateModel models in request and response do not agree")
}

func TestModelIdMatchesEmbeddedModelId(t *testing.T) {
	srv := testingServer()

	model := &api.CreateModelRequest{
		Model: &api.Model{
			ModelId: "test-model",
			Details: "This is a test",
		},
	}

	_, err := srv.CreateModel(context.Background(), model)
	assert.NoError(t, err)

	// test missing
	updateModelRequest := &api.UpdateModelRequest{
		ModelId: "test-model",
		Model: &api.Model{
			Details: "desc1",
		},
	}

	expectedModel := &api.UpdateModelResponse{
		Model: &api.Model{
			ModelId: "test-model",
			Details: "desc1",
		},
	}

	updateModelResponse, err := srv.UpdateModel(context.Background(), updateModelRequest)
	assert.NoError(t, err)
	assert.Equal(t, expectedModel, updateModelResponse)

	// test mismatch
	updateModelRequest = &api.UpdateModelRequest{
		ModelId: "test-model",
		Model: &api.Model{
			ModelId: "broken-model",
			Details: "desc1",
		},
	}

	updateModelResponse, err = srv.UpdateModel(context.Background(), updateModelRequest)
	assert.Nil(t, updateModelResponse)
	assert.Error(t, err)
}

func TestMissingModelInHyperparameterUpdate(t *testing.T) {
	srv := testingServer()

	model := &api.CreateModelRequest{
		Model: &api.Model{
			ModelId: "test-model",
			Details: "This is a test",
		},
	}

	srv.CreateModel(context.Background(), model)

	updateModelRequest := &api.UpdateModelRequest{
		ModelId: "test-model",
		Model:   nil,
	}

	updateModelResponse, err := srv.UpdateModel(context.Background(), updateModelRequest)
	assert.Nil(t, updateModelResponse)
	assert.Error(t, err)
}

func TestMissingModelIdInHyperparameterUpdate(t *testing.T) {
	srv := testingServer()

	model := &api.CreateModelRequest{
		Model: &api.Model{
			ModelId: "test-model",
			Details: "This is a test",
		},
	}

	srv.CreateModel(context.Background(), model)

	updateModelRequest := &api.UpdateModelRequest{
		Model: &api.Model{
			ModelId:                  "test-model",
			Details:                  "desc1",
			CanonicalHyperparameters: "canon1",
		},
	}

	updateModelResponse, err := srv.UpdateModel(context.Background(), updateModelRequest)
	assert.Nil(t, updateModelResponse)
	assert.Error(t, err)
}

// Creates a model and tests that GetModel returns the expected information
func TestGetModel(t *testing.T) {
	srv := testingServer()

	model := api.CreateModelRequest{
		Model: &api.Model{
			ModelId:                  "test-model",
			Details:                  "This is a test",
			CanonicalHyperparameters: "NONE",
		},
	}
	ctx := context.Background()
	_, err := srv.CreateModel(ctx, &model)
	if err != nil {
		t.Error(err)
	}

	getModelRequest := api.GetModelRequest{ModelId: model.Model.ModelId}
	getModelResponse, err := srv.GetModel(ctx, &getModelRequest)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, model.Model.ModelId, getModelResponse.ModelId, "Did not receive the expected ModelId in GetModel response")
	assert.Equal(t, model.Model.Details, getModelResponse.Details, "Did not receive the expected Details in GetModel response")
	assert.Equal(t, model.Model.CanonicalHyperparameters, getModelResponse.CanonicalHyperparameters, "Did not receive the expected CanonicalHyperparameters in GetModel response")
}

// Tests that hyperparameters are correctly created
func TestCreateAndListHyperparameters(t *testing.T) {
	srv := testingServer()

	// Create a model under which to test hyperparameters functionality
	modelID := "test-model"
	model := api.CreateModelRequest{
		Model: &api.Model{
			ModelId: modelID,
			Details: "This is a test",
		},
	}
	ctx := context.Background()
	_, err := srv.CreateModel(ctx, &model)
	if err != nil {
		t.Error(err)
	}

	createHyperparametersRequests := make([]api.CreateHyperparametersRequest, 21)
	for i := range createHyperparametersRequests {
		hyperparametersID := fmt.Sprintf("hyperparameters-%d", i)
		hyperparameters := make(map[string]string)
		hyperparameters["parameter"] = fmt.Sprintf("parameter-value-for-%d", i)
		createHyperparametersRequests[i] = api.CreateHyperparametersRequest{
			ModelId:           modelID,
			HyperparametersId: hyperparametersID,
			Hyperparameters:   hyperparameters,
		}
	}

	listHyperparametersRequest := api.ListHyperparametersRequest{
		ModelId:  modelID,
		MaxItems: int32(21),
	}

	for i, req := range createHyperparametersRequests {
		listHyperparametersResponse, err := srv.ListHyperparameters(ctx, &listHyperparametersRequest)
		if err != nil {
			t.Error(err)
		}
		if len(listHyperparametersResponse.HyperparametersIds) != i {
			t.Errorf("Incorrect number of registered hyperparameters for model %s; expected: %d, actual: %d", modelID, i, len(listHyperparametersResponse.HyperparametersIds))
		}
		createHyperparametersResponse, err := srv.CreateHyperparameters(ctx, &req)
		if err != nil {
			t.Error(err)
		}
		expectedResourcePath := fmt.Sprintf("/models/%s/hyperparameters/%s", modelID, req.HyperparametersId)
		if createHyperparametersResponse.ResourcePath != expectedResourcePath {
			t.Errorf("Incorrect resource path in CreateHyperparameters response; expected: %s, actual: %s", expectedResourcePath, createHyperparametersResponse.ResourcePath)
		}
	}
}

// Tests that hyperparameters are correctly listed (pagination behaviour)
func TestListHyperparameters(t *testing.T) {
	srv := testingServer()

	// Create a model under which to test hyperparameters functionality
	modelID := "test-model"
	model := api.CreateModelRequest{
		Model: &api.Model{
			ModelId: modelID,
			Details: "This is a test",
		},
	}
	ctx := context.Background()
	_, err := srv.CreateModel(ctx, &model)
	if err != nil {
		t.Error(err)
	}

	hpCreationRequests := make([]api.CreateHyperparametersRequest, 21)
	for i := range hpCreationRequests {
		hyperparametersID := fmt.Sprintf("hyperparameters-%d", i)
		hyperparameters := make(map[string]string)
		hyperparameters["parameter"] = fmt.Sprintf("parameter-value-for-%d", i)
		hpCreationRequests[i] = api.CreateHyperparametersRequest{
			ModelId:           modelID,
			HyperparametersId: hyperparametersID,
			Hyperparameters:   hyperparameters,
		}
	}
	hyperparametersIDs := make([]string, len(hpCreationRequests))
	for i, req := range hpCreationRequests {
		hyperparametersIDs[i] = req.HyperparametersId
		_, err := srv.CreateHyperparameters(ctx, &req)
		if err != nil {
			t.Error(err)
		}
	}
	// NOTE: HyperparametersIDs are sorted lexicographically, not chronologically!
	sort.Strings(hyperparametersIDs)

	type ListHyperparametersTest struct {
		Server                     *api.RepositoryServer
		ModelId                    string
		Marker                     string
		MaxItems                   int32
		ExpectedHyperparametersIds []string
	}

	tests := []ListHyperparametersTest{
		{
			Server:                     &srv,
			ModelId:                    modelID,
			MaxItems:                   int32(5),
			ExpectedHyperparametersIds: hyperparametersIDs[0:5],
		},
		{
			Server:                     &srv,
			ModelId:                    modelID,
			Marker:                     hyperparametersIDs[1],
			MaxItems:                   int32(5),
			ExpectedHyperparametersIds: hyperparametersIDs[2:7],
		},
		{
			Server:                     &srv,
			ModelId:                    modelID,
			Marker:                     hyperparametersIDs[15],
			MaxItems:                   int32(5),
			ExpectedHyperparametersIds: hyperparametersIDs[16:21],
		},
		{
			Server:                     &srv,
			ModelId:                    modelID,
			Marker:                     hyperparametersIDs[15],
			MaxItems:                   int32(6),
			ExpectedHyperparametersIds: hyperparametersIDs[16:21],
		},
		// TODO(frederick): Specification says that list endpoints should return items AFTER marker,
		// not after and including marker. No need to change behaviour, just make the two consistent.
		{
			Server:                     &srv,
			ModelId:                    modelID,
			Marker:                     hyperparametersIDs[0],
			MaxItems:                   int32(20),
			ExpectedHyperparametersIds: hyperparametersIDs[1:21],
		},
		{
			Server:                     &srv,
			ModelId:                    modelID,
			Marker:                     hyperparametersIDs[0],
			ExpectedHyperparametersIds: hyperparametersIDs[1:11],
		},
		{
			Server:                     &srv,
			ModelId:                    modelID,
			Marker:                     hyperparametersIDs[0],
			MaxItems:                   0,
			ExpectedHyperparametersIds: hyperparametersIDs[1:11],
		},
		{
			Server:                     &srv,
			ModelId:                    modelID,
			Marker:                     hyperparametersIDs[0],
			MaxItems:                   -10,
			ExpectedHyperparametersIds: hyperparametersIDs[1:11],
		},
	}

	for i, test := range tests {
		listHyperparametersRequest := api.ListHyperparametersRequest{
			ModelId:  test.ModelId,
			Marker:   test.Marker,
			MaxItems: test.MaxItems,
		}

		tsrv := *test.Server
		listHyperparametersResponse, err := tsrv.ListHyperparameters(ctx, &listHyperparametersRequest)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, test.ModelId, listHyperparametersResponse.ModelId)
		assert.Equalf(t, test.ExpectedHyperparametersIds, listHyperparametersResponse.HyperparametersIds, "TestListHyperparameters %d: ListHyperparameters request returned incorrect HyperparametersIds", i)

		if t.Failed() {
			break
		}
	}
}

// Tests that hyperparameters update behaviour is correct
func TestUpdateHyperparameters(t *testing.T) {
	srv := testingServer()

	// Create a model under which to test hyperparameters functionality
	modelID := "test-model"
	model := api.CreateModelRequest{
		Model: &api.Model{
			ModelId: modelID,
			Details: "This is a test",
		},
	}
	ctx := context.Background()
	_, err := srv.CreateModel(ctx, &model)
	if err != nil {
		t.Error(err)
	}

	// Create hyperparameters to set up the test
	hyperparametersID := "test-hyperparameters"
	oldHyperparameters := make(map[string]string)
	oldHyperparameters["untouched-parameter-key"] = "old-value"
	oldHyperparameters["old-parameter-key"] = "old-value"

	hpCreationRequest := api.CreateHyperparametersRequest{
		ModelId:           modelID,
		HyperparametersId: hyperparametersID,
		Hyperparameters:   oldHyperparameters,
	}
	_, err = srv.CreateHyperparameters(ctx, &hpCreationRequest)
	if err != nil {
		t.Error(err)
	}

	newHyperparameters := make(map[string]string)
	newHyperparameters["old-parameter-key"] = "new-value"
	newHyperparameters["new-parameter-key"] = "new-value"
	canonicalCheckpoint := "lol"
	hpUpdateRequest := api.UpdateHyperparametersRequest{
		ModelId:             modelID,
		HyperparametersId:   hyperparametersID,
		Hyperparameters:     newHyperparameters,
		CanonicalCheckpoint: canonicalCheckpoint,
	}
	hpUpdateResponse, err := srv.UpdateHyperparameters(ctx, &hpUpdateRequest)
	if err != nil {
		t.Error(err)
	}

	// Note: UpdateHyperparameters merges hyperparameter maps from the request value into the
	// value in storage (with the former taking precedence on conflicting keys).
	expectedHyperparameters := make(map[string]string)
	for k, v := range oldHyperparameters {
		expectedHyperparameters[k] = v
	}
	for k, v := range newHyperparameters {
		expectedHyperparameters[k] = v
	}
	assert.Equal(t, modelID, hpUpdateResponse.ModelId, "Did not receive expected ModelID in UpdateHyperparameters response")
	assert.Equal(t, hyperparametersID, hpUpdateResponse.HyperparametersId, "Did not receive expected HyperparametersID in UpdateHyperparameters response")
	assert.Equal(t, canonicalCheckpoint, hpUpdateResponse.CanonicalCheckpoint, "Did not receive expected CanonicalCheckpoint in UpdateHyperparameters response")
	assert.Equal(t, expectedHyperparameters, hpUpdateResponse.Hyperparameters, "Did not receive expected hyperparameters in UpdateHyperparameters response")
}

// Creates hyperparameters for a given model and tests that GetHyperparameters returns the expected information
func TestGetHyperparameters(t *testing.T) {
	srv := testingServer()

	// Create a model under which to test hyperparameters functionality
	modelID := "test-model"
	model := api.CreateModelRequest{
		Model: &api.Model{
			ModelId: modelID,
			Details: "This is a test",
		},
	}
	ctx := context.Background()
	_, err := srv.CreateModel(ctx, &model)
	if err != nil {
		t.Error(err)
	}

	// Create hyperparameters to set up the test
	hyperparametersID := "test-hyperparameters"
	hyperparameters := make(map[string]string)
	hyperparameters["untouched-parameter-key"] = "old-value"
	hyperparameters["old-parameter-key"] = "old-value"

	hpCreationRequest := api.CreateHyperparametersRequest{
		ModelId:           modelID,
		HyperparametersId: hyperparametersID,
		Hyperparameters:   hyperparameters,
	}
	_, err = srv.CreateHyperparameters(ctx, &hpCreationRequest)
	if err != nil {
		t.Error(err)
	}

	hpGetRequest := api.GetHyperparametersRequest{
		ModelId:           modelID,
		HyperparametersId: hyperparametersID,
	}
	hpGetResponse, err := srv.GetHyperparameters(ctx, &hpGetRequest)
	assert.NoError(t, err)
	assert.Equal(t, modelID, hpGetResponse.ModelId, "Did not receive expected ModelID in UpdateHyperparameters response")
	assert.Equal(t, hyperparametersID, hpGetResponse.HyperparametersId, "Did not receive expected HyperparametersID in UpdateHyperparameters response")
	assert.Equal(t, hyperparameters, hpGetResponse.Hyperparameters, "Did not receive expected hyperparameters in UpdateHyperparameters response")
}

// Tests that checkpoints are correctly created and listed
func TestCreateAndListCheckpoints(t *testing.T) {
	srv := testingServer()

	// Create a model and hyperparameters under which to test checkpoint functionality
	modelID := "test-model"
	model := api.CreateModelRequest{
		Model: &api.Model{
			ModelId: modelID,
			Details: "This is a test",
		},
	}
	ctx := context.Background()
	_, err := srv.CreateModel(ctx, &model)
	if err != nil {
		t.Error(err)
	}

	hyperparametersID := "test-hyperparameters"
	hyperparameters := make(map[string]string)
	hyperparameters["parameter"] = "parameter-value"

	hpCreationRequest := api.CreateHyperparametersRequest{
		ModelId:           modelID,
		HyperparametersId: hyperparametersID,
		Hyperparameters:   hyperparameters,
	}
	hpCreationResponse, err := srv.CreateHyperparameters(ctx, &hpCreationRequest)
	if err != nil {
		t.Error(err)
	}
	hyperparametersResourcePath := hpCreationResponse.ResourcePath

	ckptCreationRequests := make([]api.CreateCheckpointRequest, 21)
	for i := range ckptCreationRequests {
		ckptID := fmt.Sprintf("checkpoint-%d", i)
		link := fmt.Sprintf("http://example.com/checkpoints-for-test/%d.zip", i)
		info := make(map[string]string)
		info["parameter"] = fmt.Sprintf("value-for-%d", i)
		ckptCreationRequests[i] = api.CreateCheckpointRequest{
			ModelId:           modelID,
			HyperparametersId: hyperparametersID,
			CheckpointId:      ckptID,
			Link:              link,
			Info:              info,
		}
	}

	listCheckpointsRequest := api.ListCheckpointsRequest{
		ModelId:           modelID,
		HyperparametersId: hyperparametersID,
		MaxItems:          int32(21),
	}

	for i, req := range ckptCreationRequests {
		listCheckpointsResponse, err := srv.ListCheckpoints(ctx, &listCheckpointsRequest)
		if err != nil {
			t.Error(err)
		}
		if len(listCheckpointsResponse.CheckpointIds) != i {
			t.Errorf("Incorrect number of registered hyperparameters for model %s; expected: %d, actual: %d", modelID, i, len(listCheckpointsResponse.CheckpointIds))
		}
		createCheckpointsResponse, err := srv.CreateCheckpoint(ctx, &req)
		if err != nil {
			t.Error(err)
		}
		expectedResourcePath := fmt.Sprintf("%s/checkpoints/%s", hyperparametersResourcePath, req.CheckpointId)
		if createCheckpointsResponse.ResourcePath != expectedResourcePath {
			t.Errorf("Incorrect resource path in CreateCheckpoints response; expected: %s, actual: %s", expectedResourcePath, createCheckpointsResponse.ResourcePath)
		}
	}
}

// Tests that checkpoints are correctly listed (pagination behaviour)
func TestListCheckpoints(t *testing.T) {
	srv := testingServer()

	// Create a model and hyperparameters under which to test checkpoint functionality
	modelID := "test-model"
	model := api.CreateModelRequest{
		Model: &api.Model{
			ModelId: modelID,
			Details: "This is a test",
		},
	}
	ctx := context.Background()
	_, err := srv.CreateModel(ctx, &model)
	if err != nil {
		t.Error(err)
	}

	hyperparametersID := "test-hyperparameters"
	hyperparameters := make(map[string]string)
	hyperparameters["parameter"] = "parameter-value"

	hpCreationRequest := api.CreateHyperparametersRequest{
		ModelId:           modelID,
		HyperparametersId: hyperparametersID,
		Hyperparameters:   hyperparameters,
	}
	_, err = srv.CreateHyperparameters(ctx, &hpCreationRequest)
	if err != nil {
		t.Error(err)
	}

	ckptCreationRequests := make([]api.CreateCheckpointRequest, 21)
	for i := range ckptCreationRequests {
		checkpointID := fmt.Sprintf("checkpoint-%d", i)
		link := fmt.Sprintf("http://example.com/checkpoints-for-test/%d.zip", i)
		info := make(map[string]string)
		info["parameter"] = fmt.Sprintf("value-for-%d", i)
		ckptCreationRequests[i] = api.CreateCheckpointRequest{
			ModelId:           modelID,
			HyperparametersId: hyperparametersID,
			CheckpointId:      checkpointID,
			Link:              link,
			Info:              info,
		}
	}

	checkpointIDs := make([]string, len(ckptCreationRequests))
	for i, req := range ckptCreationRequests {
		checkpointIDs[i] = req.CheckpointId
		_, err := srv.CreateCheckpoint(ctx, &req)
		if err != nil {
			t.Error(err)
		}
	}
	// NOTE: CheckpointIds are sorted lexicographically, not chronologically!
	sort.Strings(checkpointIDs)

	type ListCheckpointsTest struct {
		Server                *api.RepositoryServer
		ModelId               string
		HyperparametersId     string
		Marker                string
		MaxItems              int32
		ExpectedCheckpointIds []string
	}

	tests := []ListCheckpointsTest{
		{
			Server:                &srv,
			ModelId:               modelID,
			HyperparametersId:     hyperparametersID,
			MaxItems:              int32(5),
			ExpectedCheckpointIds: checkpointIDs[0:5],
		},
		{
			Server:                &srv,
			ModelId:               modelID,
			HyperparametersId:     hyperparametersID,
			Marker:                checkpointIDs[1],
			MaxItems:              int32(5),
			ExpectedCheckpointIds: checkpointIDs[2:7],
		},
		{
			Server:                &srv,
			ModelId:               modelID,
			HyperparametersId:     hyperparametersID,
			Marker:                checkpointIDs[15],
			MaxItems:              int32(5),
			ExpectedCheckpointIds: checkpointIDs[16:21],
		},
		{
			Server:                &srv,
			ModelId:               modelID,
			HyperparametersId:     hyperparametersID,
			Marker:                checkpointIDs[15],
			MaxItems:              int32(6),
			ExpectedCheckpointIds: checkpointIDs[16:21],
		},
		// TODO(frederick): Specification says that list endpoints should return items AFTER marker,
		// not after and including marker. No need to change behaviour, just make the two consistent.
		{
			Server:                &srv,
			ModelId:               modelID,
			HyperparametersId:     hyperparametersID,
			Marker:                checkpointIDs[0],
			MaxItems:              int32(20),
			ExpectedCheckpointIds: checkpointIDs[1:21],
		},
		{
			Server:                &srv,
			ModelId:               modelID,
			HyperparametersId:     hyperparametersID,
			Marker:                checkpointIDs[0],
			ExpectedCheckpointIds: checkpointIDs[1:11],
		},
		{
			Server:                &srv,
			ModelId:               modelID,
			HyperparametersId:     hyperparametersID,
			Marker:                checkpointIDs[0],
			MaxItems:              0,
			ExpectedCheckpointIds: checkpointIDs[1:11],
		},
		{
			Server:                &srv,
			ModelId:               modelID,
			HyperparametersId:     hyperparametersID,
			Marker:                checkpointIDs[0],
			MaxItems:              -10,
			ExpectedCheckpointIds: checkpointIDs[1:11],
		},
	}

	for i, test := range tests {
		listCkptRequest := api.ListCheckpointsRequest{
			ModelId:           test.ModelId,
			HyperparametersId: test.HyperparametersId,
			Marker:            test.Marker,
			MaxItems:          test.MaxItems,
		}
		tsrv := *test.Server
		listCkptResponse, err := tsrv.ListCheckpoints(ctx, &listCkptRequest)
		if err != nil {
			t.Error(err)
		}
		errorMessage := fmt.Sprintf("Test %d: ListCheckpoints response does not contain the expected CheckpointIds", i)
		assert.Equalf(t, test.ExpectedCheckpointIds, listCkptResponse.CheckpointIds, errorMessage)

		assert.Equal(t, test.ModelId, listCkptResponse.ModelId)
		assert.Equal(t, test.HyperparametersId, listCkptResponse.HyperparametersId)

		if t.Failed() {
			break
		}
	}
}

// Creates a checkpoint for a given model and hyperparameters, and tests that GetCheckpoint returns the expected information
func TestGetCheckpoint(t *testing.T) {
	srv := testingServer()

	// Create a model and hyperparameters under which to test checkpoint functionality
	modelID := "test-model"
	model := api.CreateModelRequest{
		Model: &api.Model{
			ModelId: modelID,
			Details: "This is a test",
		},
	}
	ctx := context.Background()
	_, err := srv.CreateModel(ctx, &model)
	if err != nil {
		t.Error(err)
	}

	hyperparametersID := "test-hyperparameters"
	hyperparameters := make(map[string]string)
	hyperparameters["parameter"] = "parameter-value"

	hpCreationRequest := api.CreateHyperparametersRequest{
		ModelId:           modelID,
		HyperparametersId: hyperparametersID,
		Hyperparameters:   hyperparameters,
	}
	_, err = srv.CreateHyperparameters(ctx, &hpCreationRequest)
	if err != nil {
		t.Error(err)
	}

	checkpointID := "test-checkpoint"
	link := "http://example.com/checkpoints-for-test/ckpt.zip"
	info := make(map[string]string)
	info["parameter"] = "value"

	createCheckpointRequest := api.CreateCheckpointRequest{
		ModelId:           modelID,
		HyperparametersId: hyperparametersID,
		CheckpointId:      checkpointID,
		Link:              link,
		Info:              info,
	}
	_, err = srv.CreateCheckpoint(ctx, &createCheckpointRequest)
	if err != nil {
		t.Error(err)
	}

	getCheckpointRequest := api.GetCheckpointRequest{
		ModelId:           modelID,
		HyperparametersId: hyperparametersID,
		CheckpointId:      checkpointID,
	}
	getCheckpointResponse, err := srv.GetCheckpoint(ctx, &getCheckpointRequest)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, link, getCheckpointResponse.Link, "Incorrect Link in GetCheckpointResponse")

	createdAt, err := ptypes.Timestamp(getCheckpointResponse.CreatedAt)
	assert.NoError(t, err)
	assert.WithinDuration(t, time.Now(), createdAt, 2*time.Second)

	assert.Equal(t, info, getCheckpointResponse.Info, "Incorrect Info in GetCheckpointResponse")

	assert.Equal(t, modelID, getCheckpointResponse.ModelId, "Incorrect ModelId in GetCheckpointResponse")
	assert.Equal(t, hyperparametersID, getCheckpointResponse.HyperparametersId, "Incorrect HyperparametersId in GetCheckpointResponse")
	assert.Equal(t, checkpointID, getCheckpointResponse.CheckpointId, "Incorrect CheckpointId in GetCheckpointResponse")
}

func sendGetRequest(t *testing.T, url string, status int) string {
	resp, err := http.Get(url)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != status {
		t.Errorf("Expected: %d Got: %d for URL: %s", status, resp.StatusCode, url)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	return string(bodyBytes)
}

func postRequest(t *testing.T, url string, jsonStruct map[string]interface{}, status int) string {
	bytesRepresentation, err := json.Marshal(jsonStruct)
	if err != nil {
		t.Error(err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != status {
		t.Errorf("Expected: %d Got: %d for POST URL: %s", status, resp.StatusCode, url)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	return string(bodyBytes)
}

func TestIsValidID(t *testing.T) {
	assert.True(t, common.IsValidID("dii-ZZ12_"))
	assert.False(t, common.IsValidID(""))
	assert.False(t, common.IsValidID(" X"))
	assert.True(t, common.IsValidID("---"))
	assert.False(t, common.IsValidID("1&2"))
	assert.True(t, common.IsValidID("123"))
}

func TestURLEndpoints(t *testing.T) {
	storage := memory.NewMemoryRepositoryStorage()
	const grpcAddress = ":9300" // Use diff ports.
	const jsonAddress = ":9301"
	stopRequestChannel := make(chan string)
	go server.StartGrpcAndProxyServer(storage, grpcAddress, jsonAddress, stopRequestChannel)
	baseUrl := fmt.Sprintf("http://localhost%s/v1/repository/", jsonAddress)
	healthzUrl := baseUrl + "healthz"
	response := ""
	for ; response != "{\"status\":\"SERVING\"}"; response = sendGetRequest(t, healthzUrl, http.StatusOK) {
		fmt.Println("Waiting for server to become healthy")
		time.Sleep(100 * time.Millisecond)
	}
	assert.Equal(t, "{\"backendType\":\"MEMORY\"}", sendGetRequest(t, baseUrl+"config", http.StatusOK))
	assert.Equal(t, "{\"modelIds\":[]}", sendGetRequest(t, baseUrl+"models", http.StatusOK))
	const invModelErr = "\"Could not retrieve model (InvalidModelName) from storage\""
	assert.Equal(t, "{\"error\":"+invModelErr+",\"message\":"+invModelErr+",\"code\":14,\"details\":[]}",
		sendGetRequest(t, baseUrl+"models/InvalidModelName", http.StatusServiceUnavailable))
	// Seems that these requests ignore canonicalHyperparameters or any other extra tags.
	assert.Equal(t, "{\"resourcePath\":\"/models/MyModel\"}",
		postRequest(t, baseUrl+"models",
			map[string]interface{}{"model": map[string]string{
				"modelId":                  "MyModel",
				"details":                  "Selfie model",
				"canonicalHyperparameters": "batch-666",
				"randomTag":                "RandomValue",
			}}, http.StatusOK))
	assert.Equal(t, "{\"modelIds\":[\"MyModel\"]}", sendGetRequest(t, baseUrl+"models", http.StatusOK))
	assert.Equal(t, "{\"resourcePath\":\"/models/BasicModel\"}",
		postRequest(t, baseUrl+"models",
			map[string]interface{}{"model": map[string]string{
				"modelId":                  "BasicModel",
				"details":                  "Basic model",
				"canonicalHyperparameters": "batch-123",
			}}, http.StatusOK))

	// Models are sorted lexicographically, not in order of recency.
	assert.Equal(t, "{\"modelIds\":[\"BasicModel\",\"MyModel\"]}", sendGetRequest(t, baseUrl+"models", http.StatusOK))

	// This is expected to fail. One needs to create model, hyperparameters and checkpoints in sequence.
	assert.Equal(t, "Not Found\n",
		postRequest(t, baseUrl+"models/GoodModel/hyperparameterId/batch-443",
			map[string]interface{}{"model": map[string]string{
				"modelId":                  "GoodModel",
				"description":              "The best model",
				"canonicalHyperparameters": "batch-443",
			}}, http.StatusNotFound))
	assert.Equal(t, "{\"modelIds\":[\"BasicModel\",\"MyModel\"]}", sendGetRequest(t, baseUrl+"models", http.StatusOK))

	// Let's try emulating a real flow
	assert.Equal(t, "{\"resourcePath\":\"/models/MyModel/hyperparameters/HPSet1\"}",
		postRequest(t, baseUrl+"models/MyModel/hyperparameters",
			map[string]interface{}{
				"hyperparametersId": "HPSet1",
				"hyperparameters": map[string]string{
					"param1": "value1",
					"param2": "v2",
				},
			}, http.StatusOK))

	assert.Equal(t, "{\"resourcePath\":\"/models/MyModel/hyperparameters/HPSet1/checkpoints/chkpt-1\"}",
		postRequest(t, baseUrl+"models/MyModel/hyperparameters/HPSet1/checkpoints",
			map[string]interface{}{
				"checkpointId": "chkpt-1",
				"createdAt":    "1557790163",
				"info": map[string]string{
					"accuracy": "0.93",
				},
				"link": "https://example.com/h1c1.tiobundle.zip",
			}, http.StatusOK))

	assert.Equal(t, "{\"resourcePath\":\"/models/MyModel/hyperparameters/HPSet2\"}",
		postRequest(t, baseUrl+"models/MyModel/hyperparameters",
			map[string]interface{}{
				"hyperparametersId": "HPSet2",
				"hyperparameters": map[string]string{
					"number": "42",
				},
			}, http.StatusOK))

	assert.Equal(t, "{\"resourcePath\":\"/models/MyModel/hyperparameters/HPSet2/checkpoints/hp2-ckpt1\"}",
		postRequest(t, baseUrl+"models/MyModel/hyperparameters/HPSet2/checkpoints",
			map[string]interface{}{
				"checkpointId": "hp2-ckpt1",
				"createdAt":    "1557794163",
				"info": map[string]string{
					"accuracy": "0.96",
				},
				"link": "https://example.com/h2c1.tiobundle.zip",
			}, http.StatusOK))

	stopRequestChannel <- "Test Complete"
}
