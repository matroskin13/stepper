package stepper

import (
	"context"
	"encoding/json"
	"time"
)

type Context interface {
	Task() *Task
	Context() context.Context
	CreateSubtask(sub CreateTask)
	BindState(state any) error
	SetState(state any) error
	SetRetryAfter(timeout time.Duration)
}

type taskContext struct {
	ctx        context.Context
	task       *Task
	subtasks   []CreateTask
	retryAfter time.Duration

	taskEngine Engine
}

func (c *taskContext) Task() *Task {
	return c.task
}

func (c *taskContext) Context() context.Context {
	return c.ctx
}

func (c *taskContext) CreateSubtask(sub CreateTask) {
	c.subtasks = append(c.subtasks, sub)
}

func (c *taskContext) SetRetryAfter(timeout time.Duration) {
	c.retryAfter = timeout
}

func (c *taskContext) BindState(state any) error {
	if len(c.task.State) == 0 {
		return nil
	}

	return json.Unmarshal(c.task.State, state)
}

func (c *taskContext) SetState(state any) error {
	b, err := json.Marshal(state)
	if err != nil {
		return err
	}

	c.taskEngine.SetState(c.ctx, c.task.ID, b)

	return nil
}
