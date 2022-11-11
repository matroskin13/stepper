package pg

import (
	"context"
	"encoding/json"
	"time"

	"github.com/matroskin13/stepper"
)

type Task struct {
	ID               string     `json:"_id"`
	CustomId         string     `bson:"custom_id"`
	Name             string     `json:"name"`
	Data             string     `json:"data"`
	JobId            string     `json:"jobId"`
	Parent           string     `json:"parent"`
	LaunchAt         int64      `json:"launchAt"`
	Status           string     `json:"status"`
	LockAt           *time.Time `json:"lock_at"`
	State            string     `json:"state"`
	MiddlewaresState string     `json:"middlewares_state"`
	Error            *string
	EngineContext    context.Context `json:"-"`
}

func (t *Task) ToModel() *stepper.Task {

	tm := stepper.Task{
		ID:               t.ID,
		CustomId:         t.CustomId,
		Name:             t.Name,
		Data:             []byte(t.Data),
		JobId:            t.JobId,
		Parent:           t.Parent,
		LaunchAt:         time.Unix(0, t.LaunchAt),
		Status:           t.Status,
		LockAt:           t.LockAt,
		State:            []byte(t.State),
		MiddlewaresState: map[string][]byte{},
	}

	json.Unmarshal([]byte(t.MiddlewaresState), &tm.MiddlewaresState)

	return &tm
}
