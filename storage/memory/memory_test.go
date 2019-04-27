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

func TestMemory_AddHyperParameters(t *testing.T) {
	tests.Test_AddHyperParameters(t, memory.NewMemoryRepositoryStorage())
}

func TestMemory_ListHyperParameters(t *testing.T) {
	tests.Test_ListHyperParams(t, memory.NewMemoryRepositoryStorage())
}

func TestMemory_UpdateHyperParameters(t *testing.T) {
	tests.Test_UpdateHyperParams(t, memory.NewMemoryRepositoryStorage())
}

func TestMemory_AddCheckpoint(t *testing.T) {
	tests.Test_AddCheckpoint(t, memory.NewMemoryRepositoryStorage())
}

func TestMemory_ListCheckpoints(t *testing.T) {
	tests.Test_ListCheckpoints(t, memory.NewMemoryRepositoryStorage())
}
