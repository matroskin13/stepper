package stepper

import (
	"time"
)

type Task struct {
	ID               string            `json:"_id"`
	CustomId         string            `bson:"custom_id"`
	Name             string            `json:"name"`
	Data             []byte            `json:"data"`
	JobId            string            `json:"jobId"`
	Parent           string            `json:"parent"`
	LaunchAt         time.Time         `json:"launchAt"`
	Status           string            `json:"status"`
	LockAt           *time.Time        `json:"lock_at"`
	State            []byte            `json:"state"`
	MiddlewaresState map[string][]byte `json:"middlewares_state"`
}

func (t *Task) IsWaiting() bool {
	return t.Status == "waiting"
}

type CreateTask struct {
	Name        string
	Data        []byte
	CustomId    string
	LaunchAfter time.Duration
	LaunchAt    time.Time
}
