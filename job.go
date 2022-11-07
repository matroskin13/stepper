package stepper

import (
	"time"

	"github.com/robfig/cron/v3"
)

type Job struct {
	Status       string    `json:"status"`
	Name         string    `json:"name"`
	Pattern      string    `json:"pattern"`
	NextLaunchAt time.Time `json:"naxtLaunchAt"`
}

func (j *Job) CalculateNextLaunch() error {
	specParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

	schedule, err := specParser.Parse(j.Pattern)
	if err != nil {
		return err
	}

	j.NextLaunchAt = schedule.Next(time.Now())

	return nil
}

type JobConfig struct {
	Tags    []string
	Name    string
	Pattern string
}

func (c *JobConfig) NextLaunch() (time.Time, error) {
	specParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

	schedule, err := specParser.Parse(c.Pattern)
	if err != nil {
		return time.Now(), err
	}

	return schedule.Next(time.Now()), nil
}
