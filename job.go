package stepper

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/teambition/rrule-go"
)

type Job struct {
	Status        string          `json:"status"`
	Name          string          `json:"name"`
	CustomId      string          `json:"custom_id"`
	Pattern       string          `json:"pattern"`
	RRulePatern   string          `json:"rrule_pattern"`
	NextLaunchAt  time.Time       `json:"naxtLaunchAt"`
	EngineContext context.Context `json:"-"`
}

func (j *Job) CalculateNextLaunch() error {
	if j.RRulePatern != "" {
		return j.calculateNextLaunchByRRulePattern()
	}

	specParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

	schedule, err := specParser.Parse(j.Pattern)
	if err != nil {
		return err
	}

	j.NextLaunchAt = schedule.Next(time.Now())

	return nil
}

func (j *Job) calculateNextLaunchByRRulePattern() error {
	rule, err := rrule.StrToRRule(j.RRulePatern)
	if err != nil {
		return err
	}

	j.NextLaunchAt = rule.After(time.Now(), false)

	return nil
}

type JobConfig struct {
	Tags     []string
	Name     string
	Pattern  string
	CustomId string

	Schedule        *Schedule
	CalendarPattern *rrule.RRule
}

func (c *JobConfig) GetRRulePattern() (string, error) {
	rule, err := c.Schedule.toRRule()
	if err != nil {
		return "", err
	}

	return rule.String(), nil
}

func (c *JobConfig) NextLaunch() (time.Time, error) {
	if c.Schedule != nil {
		return c.getNextLaunchBySchedule()
	}

	if c.CalendarPattern != nil {
		return c.getByRRule()
	}

	specParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

	schedule, err := specParser.Parse(c.Pattern)
	if err != nil {
		return time.Now(), err
	}

	return schedule.Next(time.Now()), nil
}

func (c *JobConfig) getNextLaunchBySchedule() (time.Time, error) {
	rule, err := c.Schedule.toRRule()
	if err != nil {
		return time.Time{}, err
	}

	return rule.After(time.Now(), false), nil
}

func (c *JobConfig) getByRRule() (time.Time, error) {
	return c.CalendarPattern.After(time.Now(), false), nil
}
