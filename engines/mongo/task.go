package mongo

import (
	"time"

	"github.com/matroskin13/stepper"
)

type Task struct {
	ID               string            `bson:"id"`
	CustomId         string            `bson:"custom_id"`
	Name             string            `bson:"name"`
	Data             []byte            `bson:"data"`
	JobId            string            `bson:"jobId"`
	Parent           string            `bson:"parent"`
	LaunchAt         time.Time         `bson:"launchAt"`
	Status           string            `bson:"status"`
	LockAt           *time.Time        `bson:"lock_at"`
	State            []byte            `bson:"state"`
	MiddlewaresState map[string][]byte `bson:"middlewares_state"`
}

func (t *Task) FromModel(model *stepper.Task) {
	t.ID = model.ID
	t.CustomId = model.CustomId
	t.Name = model.Name
	t.Data = model.Data
	t.JobId = model.JobId
	t.Parent = model.Parent
	t.LaunchAt = model.LaunchAt
	t.Status = model.Status
	t.LockAt = model.LockAt
	t.State = model.State
	t.MiddlewaresState = model.MiddlewaresState
}

func (t *Task) ToModel() *stepper.Task {
	return &stepper.Task{
		ID:               t.ID,
		Name:             t.Name,
		Data:             t.Data,
		JobId:            t.JobId,
		Parent:           t.Parent,
		LaunchAt:         t.LaunchAt,
		Status:           t.Status,
		LockAt:           t.LockAt,
		State:            t.State,
		MiddlewaresState: t.MiddlewaresState,
		CustomId:         t.CustomId,
	}
}
