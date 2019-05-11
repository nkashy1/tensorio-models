package gcs_test

import (
	"github.com/doc-ai/tensorio-models/internal/tests"
	"github.com/doc-ai/tensorio-models/storage"
	"github.com/doc-ai/tensorio-models/storage/gcs"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	"testing"
)

func newTestStorage(t *testing.T, bucketName string) (storage.RepositoryStorage, *fakestorage.Server) {
	objects := make([]fakestorage.Object, 0)
	server := fakestorage.NewServer(objects)
	server.CreateBucket(bucketName)
	client := server.Client()

	repository := gcs.NewGCSStorage(client, bucketName)
	return repository, server
}

func TestGCS_AddModel(t *testing.T) {
	store, server := newTestStorage(t, "add_model")
	defer server.Stop()
	tests.Test_AddModel(t, store)
}

func TestGCS_ListModels(t *testing.T) {
	store, server := newTestStorage(t, "list_models")
	defer server.Stop()
	tests.Test_ListModels(t, store)
}

func TestGCS_UpdateModel(t *testing.T) {
	store, server := newTestStorage(t, "update_model")
	defer server.Stop()
	tests.Test_UpdateModels(t, store)
}

func TestGCS_AddHyperparameters(t *testing.T) {
	store, server := newTestStorage(t, "add_hyperparameters")
	defer server.Stop()
	tests.Test_AddHyperparameters(t, store)
}

func TestGCS_ListHyperparameters(t *testing.T) {
	store, server := newTestStorage(t, "list_hyperparameters")
	defer server.Stop()
	tests.Test_ListHyperparams(t, store)
}

func TestGCS_UpdateHyperparameters(t *testing.T) {
	store, server := newTestStorage(t, "update_hyperparameters")
	defer server.Stop()
	tests.Test_UpdateHyperparams(t, store)
}

func TestGCS_AddCheckpoint(t *testing.T) {
	store, server := newTestStorage(t, "add_checkpoint")
	defer server.Stop()
	tests.Test_AddCheckpoint(t, store)
}

func TestGCS_ListCheckpoints(t *testing.T) {
	store, server := newTestStorage(t, "list_checkpoints")
	defer server.Stop()
	tests.Test_ListCheckpoints(t, store)
}
