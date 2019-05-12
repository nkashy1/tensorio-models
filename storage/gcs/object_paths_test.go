package gcs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_objModelPath(t *testing.T) {
	modelId := "model1"
	modelPath := "model1/model.json"

	assert.Equal(t, modelPath, objModelPath(modelId))
}

func Test_objHyperparametersPath(t *testing.T) {
	modelId := "model1"
	paramId := "param2"
	modelPath := "model1/hyperparameters/param2/params.json"

	assert.Equal(t, modelPath, objHyperparametersPath(modelId, paramId))
}

func Test_objCheckpointPath(t *testing.T) {
	modelId := "model1"
	paramId := "param2"
	checkpointId := "checkpoint3"
	modelPath := "model1/hyperparameters/param2/checkpoints/checkpoint3/checkpoint.json"

	assert.Equal(t, modelPath, objCheckpointPath(modelId, paramId, checkpointId))
}

func Test_extractModelName(t *testing.T) {
	path := "model1/"
	name := extractObjectName(path)
	assert.Equal(t, name, "model1")
}

func Test_extractHyperparametersName(t *testing.T) {
	path := "model1/hyperparameters/param2/"
	name := extractObjectName(path)
	assert.Equal(t, name, "param2")
}

func Test_extractCheckpointName(t *testing.T) {
	path := "model1/hyperparameters/param2/checkpoints/checkpoint3/"
	name := extractObjectName(path)
	assert.Equal(t, name, "checkpoint3")
}
