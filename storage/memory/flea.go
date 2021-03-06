package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/doc-ai/tensorio-models/api"
	"github.com/doc-ai/tensorio-models/common"
	"github.com/doc-ai/tensorio-models/storage"
	"github.com/google/uuid"
)

type flea struct {
	lock              *sync.RWMutex
	tasks             map[string]storage.Task
	repositoryBaseURL string
	uploadReqURL      string
}

// NewMemoryFleaStorage - returns in-memory implementation of FleaStorage interface.
func NewMemoryFleaStorage(repositoryBaseURL string) storage.FleaStorage {
	store := &flea{
		lock:              &sync.RWMutex{},
		repositoryBaseURL: repositoryBaseURL,
		uploadReqURL:      "gs://example-repo", // Stub in this implementation.
		tasks:             make(map[string]storage.Task),
	}
	return store
}

func (s *flea) GetStorageType() string { return "MEMORY" }

func (s *flea) GetBucketName() string { return "" }

func (s *flea) AddTask(ctx context.Context, req api.TaskDetails) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, exists := s.tasks[req.TaskId]
	if exists {
		return storage.ErrDuplicateTaskId
	}
	s.tasks[req.TaskId] = storage.Task{
		ModelId:           req.ModelId,
		HyperparametersId: req.HyperparametersId,
		CheckpointId:      req.CheckpointId,
		TaskId:            req.TaskId,
		Deadline:          req.Deadline,
		Active:            req.Active,
		Link:              req.Link,
		CreatedTime:       time.Now(),
		Jobs:              make(map[string]storage.Job),
	}
	return nil
}

func (s *flea) ModifyTask(ctx context.Context, req api.ModifyTaskRequest) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	task, exists := s.tasks[req.TaskId]
	if !exists {
		return storage.ErrMissingTaskId
	}
	task.Deadline = req.Deadline
	task.Active = req.Active
	s.tasks[req.TaskId] = task
	return nil
}

// Expects that the input sanity checks are done by the caller.
func (s *flea) ListTasks(ctx context.Context, req api.ListTasksRequest) (api.ListTasksResponse, error) {
	resp := api.ListTasksResponse{}
	isLimited := false
	if req.MaxItems > 0 {
		isLimited = true
		resp.MaxItems = req.MaxItems
	}
	resp.StartTaskId = req.StartTaskId
	s.lock.RLock()
	defer s.lock.RUnlock()
	var taskIds []string
	for taskId, task := range s.tasks {
		if !req.IncludeInactive && !task.Active {
			continue
		}
		if req.StartTaskId != "" && taskId < req.StartTaskId {
			continue
		}
		if task.ModelId != req.ModelId && req.ModelId != "" {
			continue
		}
		if task.HyperparametersId != req.HyperparametersId && req.HyperparametersId != "" {
			continue
		}
		if task.CheckpointId != req.CheckpointId && req.CheckpointId != "" {
			continue
		}
		taskIds = append(taskIds, taskId)
	}
	sort.Strings(taskIds)
	if isLimited && (len(taskIds) >= int(req.MaxItems)) {
		taskIds = taskIds[:req.MaxItems]
	}
	resp.TaskIds = taskIds
	return resp, nil
}

func (s *flea) GetTask(ctx context.Context, taskId string) (api.TaskDetails, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	task, exists := s.tasks[taskId]
	if !exists {
		return api.TaskDetails{}, storage.ErrMissingTaskId
	}
	resp := api.TaskDetails{
		ModelId:           task.ModelId,
		HyperparametersId: task.HyperparametersId,
		CheckpointId:      task.CheckpointId,
		TaskId:            task.TaskId,
		Deadline:          task.Deadline,
		Active:            task.Active,
		Link:              task.Link,
		CheckpointLink: s.repositoryBaseURL + common.GetCheckpointResourcePath(
			task.ModelId, task.HyperparametersId, task.CheckpointId),
	}
	return resp, nil
}

func (s *flea) StartTask(ctx context.Context, taskId string) (api.StartTaskResponse, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	task, exists := s.tasks[taskId]
	if !exists {
		return api.StartTaskResponse{}, storage.ErrMissingTaskId
	}
	jobId := uuid.New().String()
	uploadTo := fmt.Sprintf("%s/tasksJobs/%s/%s.zip", s.uploadReqURL, taskId, jobId)
	resp := api.StartTaskResponse{
		Status:   api.StartTaskResponse_APPROVED,
		JobId:    jobId,
		UploadTo: uploadTo,
	}
	task.Jobs[resp.JobId] = storage.Job{
		TaskId:       taskId,
		JobId:        jobId,
		UploadUrl:    uploadTo,
		AcceptedTime: time.Now(),
		Errors:       make([]string, 0),
	}
	s.tasks[taskId] = task
	return resp, nil
}

func (s *flea) AddJobError(ctx context.Context, req api.JobErrorRequest) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	task, exists := s.tasks[req.TaskId]
	if !exists {
		return storage.ErrMissingTaskId
	}
	job, exists := task.Jobs[req.JobId]
	if !exists {
		return storage.ErrMissingJobId
	}
	job.Errors = append(job.Errors, req.ErrorMessage)
	task.Jobs[req.JobId] = job
	s.tasks[req.TaskId] = task
	return nil
}
