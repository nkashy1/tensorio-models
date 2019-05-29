package storage

import (
	"context"
	"errors"
	"time"

	"github.com/doc-ai/tensorio-models/api"
	"github.com/golang/protobuf/ptypes/timestamp"
)

var ModelDoesNotExistError = errors.New("Model does not exist")
var ModelExistsError = errors.New("Model already exists")

var HyperparametersDoesNotExistError = errors.New("Hyperparameters does not exist")
var HyperparametersExistsError = errors.New("Hyperparameters already exists")

var CheckpointDoesNotExistError = errors.New("Checkpoint does not exist")
var CheckpointExistsError = errors.New("Checkpoint already exists")

type Model struct {
	ModelId                  string
	Details                  string
	CanonicalHyperparameters string
}

type Hyperparameters struct {
	ModelId             string
	HyperparametersId   string
	CanonicalCheckpoint string
	UpgradeTo           string
	Hyperparameters     map[string]string
}

type Checkpoint struct {
	ModelId           string
	HyperparametersId string
	CheckpointId      string
	Link              string
	CreatedAt         time.Time
	Info              map[string]string
}

type RepositoryStorage interface {
	GetStorageType() string
	GetBucketName() string

	// MODELS

	ListModels(ctx context.Context, marker string, maxItems int) ([]string, error)
	GetModel(ctx context.Context, modelId string) (Model, error)

	AddModel(ctx context.Context, model Model) error
	UpdateModel(ctx context.Context, model Model) (Model, error)

	// HYPERPARAMETERS

	ListHyperparameters(ctx context.Context, modelId, marker string, maxItems int) ([]string, error)
	GetHyperparameters(ctx context.Context, modelId string, hyperparametersId string) (Hyperparameters, error)

	AddHyperparameters(ctx context.Context, hyperparameters Hyperparameters) error
	UpdateHyperparameters(ctx context.Context, hyperparameters Hyperparameters) (Hyperparameters, error)

	// CHECKPOINTS

	ListCheckpoints(ctx context.Context, modelId, hyperparametersId, marker string, maxItems int) ([]string, error)
	GetCheckpoint(ctx context.Context, modelId, hyperparametersId, checkpointId string) (Checkpoint, error)

	AddCheckpoint(ctx context.Context, checkpoint Checkpoint) error
}

type Job struct {
	TaskId    string
	JobId     string
	UploadUrl string

	ClientId     string // Get it from AuthToken?
	AcceptedTime time.Time
}

type Task struct {
	ModelId           string
	HyperparametersId string
	CheckpointId      string
	TaskId            string
	Deadline          *timestamp.Timestamp
	Active            bool
	Link              string
	CreatedTime       time.Time
	Jobs              map[string]Job // Map from JobId to Job detail
}

var ErrDuplicateTaskId = errors.New("TaskId already exists")
var ErrMissingTaskId = errors.New("Missing TaskId")
var ErrMissingJobId = errors.New("Missing JobId")
var ErrMissingModelId = errors.New("Missing ModelId")
var ErrMissingHyperparametersId = errors.New("Missing HyperparametersId")
var ErrMissingCheckpointId = errors.New("Missing CheckpointId")
var ErrInvalidModelHyperparamsCheckpointCombo = errors.New("CheckpointId requires HyperparamsId and HyperparamsId requires ModelId")
var ErrInvalidTaskId = errors.New("Invalid TaskId")
var ErrInvalidJobId = errors.New("Invalid JobId")
var ErrInvalidModelId = errors.New("Invalid ModelId")
var ErrInvalidHyperparametersId = errors.New("Invalid HyperparametersId")
var ErrInvalidCheckpointId = errors.New("Invalid CheckpointId")

type FleaStorage interface {
	GetStorageType() string
	GetBucketName() string

	AddTask(ctx context.Context, req api.TaskDetails) error
	ModifyTask(ctx context.Context, req api.ModifyTaskRequest) error

	ListTasks(ctx context.Context, req api.ListTasksRequest) (resp api.ListTasksResponse, e error)
	GetTask(ctx context.Context, taskId string) (api.TaskDetails, error)
	StartTask(ctx context.Context, taskId string) (api.StartTaskResponse, error)
}
