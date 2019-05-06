package memory_test

import (
	"github.com/doc-ai/tensorio-models/internal/tests"
	"github.com/doc-ai/tensorio-models/storage/memory"
	"testing"
)

func TestMemory_AddModel(t *testing.T) {
	tests.Test_AddModel(t, memory.NewMemoryRepositoryStorage())
}

func TestMemory_ListModels(t *testing.T) {
	tests.Test_ListModels(t, memory.NewMemoryRepositoryStorage())
}

func TestMemory_UpdateModel(t *testing.T) {
	tests.Test_UpdateModels(t, memory.NewMemoryRepositoryStorage())
}

func TestMemory_AddHyperparameters(t *testing.T) {
	tests.Test_AddHyperparameters(t, memory.NewMemoryRepositoryStorage())
}

func TestMemory_ListHyperparameters(t *testing.T) {
	tests.Test_ListHyperparams(t, memory.NewMemoryRepositoryStorage())
}

func TestMemory_UpdateHyperparameters(t *testing.T) {
	tests.Test_UpdateHyperparams(t, memory.NewMemoryRepositoryStorage())
}

func TestMemory_AddCheckpoint(t *testing.T) {
	tests.Test_AddCheckpoint(t, memory.NewMemoryRepositoryStorage())
}

func TestMemory_ListCheckpoints(t *testing.T) {
	tests.Test_ListCheckpoints(t, memory.NewMemoryRepositoryStorage())
}
