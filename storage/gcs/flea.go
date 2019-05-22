package gcs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	gcs "cloud.google.com/go/storage"
	"github.com/doc-ai/tensorio-models/api"
	"github.com/doc-ai/tensorio-models/common"
	"github.com/doc-ai/tensorio-models/storage"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

type flea struct {
	client             *gcs.Client
	bucket             *gcs.BucketHandle
	repositoryBaseURL  string
	uploadToBucketName string
}

// GenerateNewFleaGCSStorageFromEnv - Uses the GOOGLE_APPLICATION_CREDENTIALS and FLEA_GCS_BUCKET
// environment variables to instantiate a GCS Storage backend for tensorio-models repository
func GenerateNewFleaGCSStorageFromEnv(repositoryBaseURL string) storage.FleaStorage {
	bucketName := os.Getenv("FLEA_GCS_BUCKET")
	if bucketName == "" {
		err := errors.New("FLEA_GCS_BUCKET environment variable not defined")
		panic(err)
	}

	uploadBucketName := os.Getenv("FLEA_UPLOAD_GCS_BUCKET")
	if uploadBucketName == "" {
		err := errors.New("FLEA_UPLOAD_GCS_BUCKET environment variable not set")
		panic(err)
	}

	ctx := context.Background()
	gcsClient, err := gcs.NewClient(ctx)
	if err != nil {
		panic(err)
	}
	bucket := gcsClient.Bucket(bucketName)

	return &flea{client: gcsClient,
		bucket:             bucket,
		repositoryBaseURL:  repositoryBaseURL,
		uploadToBucketName: uploadBucketName}
}

func (store flea) GetStorageType() string {
	return "GCS"
}

func objTaskPath(taskId string) string {
	return "tasks/" + taskId + "/task.json"
}

func (store flea) GetUploadToURL(taskId, jobId string) string {
	return fmt.Sprintf("gs://%s/tasksJobs/%s/%s.zip", store.uploadToBucketName, taskId, jobId)
}

func (store flea) AddTask(ctx context.Context, req api.TaskDetails) error {
	objLoc := objTaskPath(req.TaskId)
	object := store.bucket.Object(objLoc)

	_, err := object.Attrs(ctx)
	if err != gcs.ErrObjectNotExist {
		if err == nil {
			return storage.ErrDuplicateTaskId
		}
		return err
	}

	writer := object.NewWriter(ctx)
	bytes, err := json.Marshal(req)
	return writeObject(ctx, writer, bytes)
}

func (store flea) GetTask(ctx context.Context, taskId string) (api.TaskDetails, error) {
	task := api.TaskDetails{}
	objLoc := objTaskPath(taskId)
	object := store.bucket.Object(objLoc)
	reader, err := object.NewReader(ctx)
	if err != nil {
		return task, storage.ErrMissingTaskId
	}

	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return task, err
	}
	err = json.Unmarshal(bytes, &task)
	return task, err
}

func (store flea) ModifyTask(ctx context.Context, req api.ModifyTaskRequest) error {
	task, err := store.GetTask(ctx, req.TaskId)
	if err != nil {
		return err
	}
	task.Deadline = req.Deadline
	task.Active = req.Active

	objLoc := objTaskPath(req.TaskId)
	object := store.bucket.Object(objLoc)
	writer := object.NewWriter(ctx)
	bytes, err := json.Marshal(task)
	return writeObject(ctx, writer, bytes)
}

func (store flea) StartTask(ctx context.Context, taskId string) (api.StartTaskResponse, error) {
	resp := api.StartTaskResponse{}
	_, err := store.GetTask(ctx, taskId)
	if err != nil {
		return resp, err
	}
	resp.JobId = uuid.New().String()
	resp.UploadTo = store.GetUploadToURL(taskId, resp.JobId)
	resp.Status = api.StartTaskResponse_APPROVED
	return resp, nil
}

func (store flea) ListTasks(ctx context.Context, req api.ListTasksRequest) (api.ListTasksResponse, error) {
	resp := api.ListTasksResponse{}
	query := &gcs.Query{
		Delimiter: "/",
		Prefix:    "tasks/",
		Versions:  false,
	}
	iter := store.bucket.Objects(ctx, query)
	resp.Tasks = make(map[string]string)
	for {
		if req.MaxItems > 0 && len(resp.Tasks) == int(req.MaxItems) {
			break
		}
		obj, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return resp, err
		}

		taskId := extractObjectName(obj.Prefix)

		if req.StartTaskId != "" && taskId < req.StartTaskId {
			continue
		}

		task, err := store.GetTask(ctx, taskId)
		if err != nil {
			return resp, errors.New("Could not get task: " + taskId)
		}
		resp.Tasks[taskId] =
			common.GetCheckpointResourcePath(
				task.ModelId, task.HyperparametersId, task.CheckpointId)
	}
	resp.StartTaskId = req.StartTaskId
	resp.MaxItems = req.MaxItems
	resp.RepositoryBaseUrl = store.repositoryBaseURL
	return resp, nil
}