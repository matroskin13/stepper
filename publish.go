package stepper

import "time"

type PublishOption func(c *CreateTask)

func SetDelay(d time.Duration) PublishOption {
	return func(c *CreateTask) {
		c.LaunchAfter = d
	}
}

func LaunchAt(t time.Time) PublishOption {
	return func(c *CreateTask) {
		c.LaunchAt = t
	}
}
