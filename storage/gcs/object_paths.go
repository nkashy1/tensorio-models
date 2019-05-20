package gcs

import (
	"fmt"
	"strings"
)

func objModelPath(modelId string) string {
	objLoc := fmt.Sprintf("models/%s/model.json", modelId)
	return objLoc
}

func objHyperparametersPath(modelId string, hyperparametersId string) string {
	objLoc := fmt.Sprintf("models/%s/hyperparameters/%s/params.json", modelId, hyperparametersId)
	return objLoc
}

func objCheckpointPath(modelId string, hyperparametersId string, checkpointId string) string {
	objLoc := fmt.Sprintf("models/%s/hyperparameters/%s/checkpoints/%s/checkpoint.json", modelId, hyperparametersId, checkpointId)
	return objLoc
}

func extractObjectName(name string) string {
	splitNames := strings.Split(name, "/")
	name = splitNames[len(splitNames)-2]
	return name
}
