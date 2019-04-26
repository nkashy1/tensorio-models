package storage

import (
	"context"
	"time"
)

type Model struct {
	modelId                  string
	description              string
	canonicalHyperParameters string
}

type HyperParameters struct {
	ModelId             string
	HyperParametersId   string
	CanonicalCheckpoint string
	HyperParameters     map[string]string
}

type Checkpoint struct {
	ModelId           string
	HyperParametersId string
	CheckpointId      string
	Link              string
	CreatedAt         time.Time
	Info              map[string]string
}

type RepositoryStorage interface {
	// MODELS

	ListModels(ctx context.Context, marker string, maxItems int) ([]string, error)
	GetModel(ctx context.Context, modelId string) (Model, error)

	AddModel(ctx context.Context, model Model) error
	UpdateModel(ctx context.Context, model Model) (Model, error)

	// HYPERPARAMETERS

	ListHyperParameters(ctx context.Context, modelId, marker string, maxItems int) ([]string, error)
	GetHyperparameters(ctx context.Context, modelId string, hyperParametersId string) (HyperParameters, error)

	AddHyperParameters(ctx context.Context, hyperParameters HyperParameters) error
	UpdateHyperParameters(ctx context.Context, hyperParameters HyperParameters) (HyperParameters, error)

	// CHECKPOINTS

	ListCheckpoints(ctx context.Context, modelId, hyperParametersId, marker string, maxItems int) ([]string, error)
	GetCheckpoint(ctx context.Context, modelId, hyperParametersId, checkpointId string) (Checkpoint, error)

	AddCheckpoint(ctx context.Context, checkpoint Checkpoint) error
	UpdateCheckpoint(ctx context.Context, checkpoint Checkpoint) (Checkpoint, error)
}
