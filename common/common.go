package common

import (
	"fmt"
	"strings"
)

// IsValidIDChar - returns whether c is a valid character - alphanumeric, _ and -
func IsValidIDChar(c byte) bool {
	if (c >= 'A') && (c <= 'Z') {
		return true
	}
	if (c >= 'a') && (c <= 'z') {
		return true
	}
	if (c >= '0') && (c <= '9') {
		return true
	}
	if (c == '-') || (c == '_') {
		return true
	}
	return false
}

// IsValidID - returns whether s is non-empty string of valid characters
func IsValidID(s string) bool {
	l := len(s)
	for i := 0; i < l; i++ {
		if !IsValidIDChar(s[i]) {
			return false
		}
	}
	return s != ""
}

func GetCheckpointResourcePath(modelID, hyperparametersID, checkpointID string) string {
	resourcePath := fmt.Sprintf("/models/%s/hyperparameters/%s/checkpoints/%s", modelID, hyperparametersID, checkpointID)
	return resourcePath
}

// RepositoryStorage implementations return resources in the form:
// <modelId>, <modelId>:<hyperparametersId>, <modelId>:<hyperparametersId>:<checkpointId>
// This function takes input in those formats and returns (respectively):
// <modelId>, <hyperparametersId>, <checkpointId>
func GetTerminalResourceFromStoragePath(storagePath string) string {
	storageDelimiter := ":"
	components := strings.Split(storagePath, storageDelimiter)
	terminalComponent := components[len(components)-1]
	return terminalComponent
}
