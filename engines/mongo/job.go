package mongo

import (
	"time"

	"github.com/matroskin13/stepper"
)

type job struct {
	Status       string    `bson:"status"`
	Name         string    `bson:"name"`
	Pattern      string    `bson:"pattern"`
	NextLaunchAt time.Time `bson:"naxtLaunchAt"`
	CustomId     string    `bson:"custom_id"`
	RRulePatern  string    `bson:"rrule_pattern"`
}

func (j *job) FromModel(model *stepper.Job) {
	j.Status = model.Status
	j.Name = model.Name
	j.Pattern = model.Pattern
	j.NextLaunchAt = model.NextLaunchAt
	j.CustomId = model.CustomId
	j.RRulePatern = model.RRulePatern
}

func (j *job) ToModel() *stepper.Job {
	return &stepper.Job{
		Status:       j.Status,
		Name:         j.Name,
		Pattern:      j.Pattern,
		NextLaunchAt: j.NextLaunchAt,
		CustomId:     j.CustomId,
		RRulePatern:  j.RRulePatern,
	}
}
