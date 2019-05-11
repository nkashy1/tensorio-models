package gcs

import (
	gcs "cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"fmt"
	"github.com/doc-ai/tensorio-models/storage"
	"google.golang.org/api/iterator"
	"io"
	"io/ioutil"
	"strings"
)

type gcsStorage struct {
	client *gcs.Client
	bucket *gcs.BucketHandle
}

func NewGCSStorage(client *gcs.Client, bucketName string) storage.RepositoryStorage {
	res := &gcsStorage{}
	res.client = client

	bucket := client.Bucket(bucketName)
	res.bucket = bucket

	return res
}

func (store gcsStorage) ListModels(ctx context.Context, marker string, maxItems int) ([]string, error) {
	query := &gcs.Query{
		Delimiter: "/",
		Prefix:    "",
		Versions:  false,
	}
	iter := store.bucket.Objects(ctx, query)

	res, err := listObjects(maxItems, iter, marker)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (store gcsStorage) GetModel(ctx context.Context, modelId string) (storage.Model, error) {
	objLoc := objModelPath(modelId)
	object := store.bucket.Object(objLoc)
	reader, err := object.NewReader(ctx)
	if err != nil {
		return storage.Model{}, err
	}

	// TODO this is dangerous, we should change this eventually to read a limited amount of data
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		if err == gcs.ErrObjectNotExist {
			return storage.Model{}, storage.ModelDoesNotExistError
		}
		return storage.Model{}, err
	}

	model := storage.Model{}

	err = json.Unmarshal(bytes, &model)
	if err != nil {
		return storage.Model{}, err
	}

	return model, nil
}

func (store gcsStorage) AddModel(ctx context.Context, model storage.Model) error {
	objLoc := objModelPath(model.ModelId)
	object := store.bucket.Object(objLoc)

	_, err := object.Attrs(ctx)
	if err != gcs.ErrObjectNotExist {
		if err == nil {
			return storage.ModelExistsError
		}
		return err
	}

	writer := object.NewWriter(ctx)
	bytes, err := json.Marshal(model)

	err = writeObject(ctx, writer, bytes)
	if err != nil {
		return err
	}

	return nil
}

func (store gcsStorage) UpdateModel(ctx context.Context, model storage.Model) (storage.Model, error) {
	objLoc := objModelPath(model.ModelId)
	object := store.bucket.Object(objLoc)

	storedModel, err := store.GetModel(ctx, model.ModelId)
	if err != nil {
		return storage.Model{}, err
	}

	if strings.TrimSpace(model.CanonicalHyperparameters) != "" {
		storedModel.CanonicalHyperparameters = model.CanonicalHyperparameters
	}
	if strings.TrimSpace(model.Details) != "" {
		storedModel.Details = model.Details
	}

	bytes, err := json.Marshal(storedModel)
	if err != nil {
		return storage.Model{}, err
	}

	writer := object.NewWriter(ctx)

	err = writeObject(ctx, writer, bytes)
	if err != nil {
		return storage.Model{}, err
	}

	return storedModel, nil
}

func (store gcsStorage) ListHyperparameters(ctx context.Context, modelId, marker string, maxItems int) ([]string, error) {
	_, err := store.GetModel(ctx, modelId)
	if err != nil {
		return nil, err
	}

	query := &gcs.Query{
		Delimiter: "/",
		Prefix:    fmt.Sprintf("%s/hyperparameters/", modelId),
		Versions:  false,
	}
	iter := store.bucket.Objects(ctx, query)

	res, err := listObjects(maxItems, iter, marker)

	for i, name := range res {
		res[i] = fmt.Sprintf("%s:%s", modelId, name)
	}

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (store gcsStorage) GetHyperparameters(ctx context.Context, modelId string, hyperparametersId string) (storage.Hyperparameters, error) {
	objLoc := objParamPath(modelId, hyperparametersId)
	object := store.bucket.Object(objLoc)

	_, err := store.GetModel(ctx, modelId)
	if err != nil {
		return storage.Hyperparameters{}, err
	}
	reader, err := object.NewReader(ctx)
	if err != nil {
		if err == gcs.ErrObjectNotExist {
			return storage.Hyperparameters{}, storage.HyperparametersDoesNotExistError
		}
		return storage.Hyperparameters{}, err
	}

	// TODO this is dangerous, we should change this eventually to read a limited amount of data
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return storage.Hyperparameters{}, err
	}

	hyperparameters := storage.Hyperparameters{}

	err = json.Unmarshal(bytes, &hyperparameters)
	if err != nil {
		return storage.Hyperparameters{}, err
	}

	return hyperparameters, nil
}

func (store gcsStorage) AddHyperparameters(ctx context.Context, hyperparameters storage.Hyperparameters) error {
	objLoc := objParamPath(hyperparameters.ModelId, hyperparameters.HyperparametersId)
	object := store.bucket.Object(objLoc)

	_, err := store.GetModel(ctx, hyperparameters.ModelId)
	if err != nil {
		return err
	}

	_, err = object.Attrs(ctx)
	if err != gcs.ErrObjectNotExist {
		if err == nil {
			return storage.ModelExistsError
		}
		return err
	}

	writer := object.NewWriter(ctx)
	bytes, err := json.Marshal(hyperparameters)

	err = writeObject(ctx, writer, bytes)
	if err != nil {
		return err
	}

	return nil
}

func (store gcsStorage) UpdateHyperparameters(ctx context.Context, hyperparameters storage.Hyperparameters) (storage.Hyperparameters, error) {
	objLoc := objParamPath(hyperparameters.ModelId, hyperparameters.HyperparametersId)
	object := store.bucket.Object(objLoc)

	storedHyperparameters, err := store.GetHyperparameters(ctx, hyperparameters.ModelId, hyperparameters.HyperparametersId)
	if err != nil {
		return storage.Hyperparameters{}, err
	}

	if strings.TrimSpace(hyperparameters.CanonicalCheckpoint) != "" {
		storedHyperparameters.CanonicalCheckpoint = hyperparameters.CanonicalCheckpoint
	}
	if strings.TrimSpace(hyperparameters.UpgradeTo) != "" {
		storedHyperparameters.UpgradeTo = hyperparameters.UpgradeTo
	}

	if hyperparameters.Hyperparameters != nil {
		for k, v := range hyperparameters.Hyperparameters {
			storedHyperparameters.Hyperparameters[k] = v
		}
	}

	bytes, err := json.Marshal(storedHyperparameters)
	if err != nil {
		return storage.Hyperparameters{}, err
	}

	writer := object.NewWriter(ctx)

	err = writeObject(ctx, writer, bytes)
	if err != nil {
		return storage.Hyperparameters{}, err
	}

	return storedHyperparameters, nil
}

func (store gcsStorage) ListCheckpoints(ctx context.Context, modelId, hyperparametersId, marker string, maxItems int) ([]string, error) {
	_, err := store.GetModel(ctx, modelId)
	if err != nil {
		return nil, err
	}

	_, err = store.GetHyperparameters(ctx, modelId, hyperparametersId)
	if err != nil {
		return nil, err
	}

	query := &gcs.Query{
		Delimiter: "/",
		Prefix:    fmt.Sprintf("%s/hyperparameters/%s/checkpoints/", modelId, hyperparametersId),
		Versions:  false,
	}
	iter := store.bucket.Objects(ctx, query)

	res, err := listObjects(maxItems, iter, marker)

	for i, name := range res {
		res[i] = fmt.Sprintf("%s:%s:%s", modelId, hyperparametersId, name)
	}

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (store gcsStorage) GetCheckpoint(ctx context.Context, modelId, hyperparametersId, checkpointId string) (storage.Checkpoint, error) {
	objLoc := objCheckpointPath(modelId, hyperparametersId, checkpointId)

	_, err := store.GetModel(ctx, modelId)
	if err != nil {
		return storage.Checkpoint{}, err
	}
	object := store.bucket.Object(objLoc)
	reader, err := object.NewReader(ctx)
	if err != nil {
		if err == gcs.ErrObjectNotExist {
			return storage.Checkpoint{}, storage.CheckpointDoesNotExistError
		}
		return storage.Checkpoint{}, err
	}

	// TODO this is dangerous, we should change this eventually to read a limited amount of data
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return storage.Checkpoint{}, err
	}

	checkpoint := storage.Checkpoint{}

	err = json.Unmarshal(bytes, &checkpoint)
	if err != nil {
		return storage.Checkpoint{}, err
	}

	return checkpoint, nil

}

func (store gcsStorage) AddCheckpoint(ctx context.Context, checkpoint storage.Checkpoint) error {
	objLoc := objCheckpointPath(checkpoint.ModelId, checkpoint.HyperparametersId, checkpoint.CheckpointId)
	object := store.bucket.Object(objLoc)

	_, err := store.GetModel(ctx, checkpoint.ModelId)
	if err != nil {
		return err
	}

	_, err = store.GetHyperparameters(ctx, checkpoint.ModelId, checkpoint.HyperparametersId)
	if err != nil {
		return err
	}

	_, err = object.Attrs(ctx)
	if err != gcs.ErrObjectNotExist {
		if err == nil {
			return storage.ModelExistsError
		}
		return err
	}

	writer := object.NewWriter(ctx)
	bytes, err := json.Marshal(checkpoint)

	err = writeObject(ctx, writer, bytes)
	if err != nil {
		return err
	}

	return nil
}

func writeObject(ctx context.Context, writer io.WriteCloser, bytes []byte) error {

	written := 0
	for written < len(bytes) {
		n, err := writer.Write(bytes[written:])
		written += n
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
		}
	}

	if err := writer.Close(); err != nil {
		return err
	}
	return nil
}

func listObjects(maxItems int, iter *gcs.ObjectIterator, marker string) ([]string, error) {
	res := make([]string, 0)
	for {
		if len(res) == maxItems {
			break
		}
		obj, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}

		name := obj.Prefix
		splitNames := strings.Split(name, "/")
		name = splitNames[len(splitNames)-2]

		if name <= marker {
			continue
		}

		res = append(res, name)
	}
	return res, nil
}
