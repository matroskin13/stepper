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
	GetRelatedTask(ctx context.Context, task *Task) (*Task, error)
	FindNextTask(ctx context.Context, statuses []string) (*Task, error)
	ReleaseTask(ctx context.Context, task *Task) error
	WaitTaskForSubtasks(ctx context.Context, task *Task) error
	FailTask(ctx context.Context, task *Task, err error, timeout time.Duration) error
	CreateTask(ctx context.Context, task *Task) error
	GetUnreleasedTaskChildren(ctx context.Context, task *Task) (*Task, error)
	SetState(ctx context.Context, task *Task, state []byte) error
}

type JobEngine interface {
	FindNextJob(ctx context.Context, statuses []string) (*Job, error)
	GetUnreleasedJobChildren(ctx context.Context, name string) (*Task, error)
	Release(ctx context.Context, job *Job, nextLaunchAt time.Time) error
	WaitForSubtasks(ctx context.Context, job *Job) error
	RegisterJob(ctx context.Context, cfg *JobConfig) error
	Init(ctx context.Context) error
}
