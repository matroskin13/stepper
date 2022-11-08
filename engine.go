package stepper

import (
	"context"
	"time"
)

type Engine interface {
	TaskEngine
	JobEngine
}

type TaskEngine interface {
	GetRelatedTask(ctx context.Context, task string, id string) (*Task, error)
	FindNextTask(ctx context.Context, statuses []string) (*Task, error)
	ReleaseTask(ctx context.Context, id string) error
	WaitTaskForSubtasks(ctx context.Context, id string) error
	FailTask(ctx context.Context, id string, err error, timeout time.Duration) error
	CreateTask(ctx context.Context, task *Task) error
	GetUnreleasedTaskChildren(ctx context.Context, id string) (*Task, error)
	SetState(ctx context.Context, id string, state []byte) error
}

type JobEngine interface {
	FindNextJob(ctx context.Context, statuses []string) (*Job, error)
	GetUnreleasedJobChildren(ctx context.Context, jobName string) (*Task, error)
	Release(ctx context.Context, jobName string, nextLaunchAt time.Time) error
	WaitForSubtasks(ctx context.Context, jobName string) error
	RegisterJob(ctx context.Context, cfg *JobConfig) error
}
