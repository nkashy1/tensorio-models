package gcs

import "fmt"

func objModelPath(modelId string) string {
	objLoc := fmt.Sprintf("%s/model.json", modelId)
	return objLoc
}

func objParamPath(modelId string, hyperparametersId string) string {
	objLoc := fmt.Sprintf("%s/hyperparameters/%s/params.json", modelId, hyperparametersId)
	return objLoc
}

func objCheckpointPath(modelId string, hyperparametersId string, checkpointId string) string {
	objLoc := fmt.Sprintf("%s/hyperparameters/%s/checkpoints/%s/checkpoint.json", modelId, hyperparametersId, checkpointId)
	return objLoc
}
