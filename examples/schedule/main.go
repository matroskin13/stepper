package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/matroskin13/stepper"
	mongoEngine "github.com/matroskin13/stepper/engines/mongo"
	"github.com/matroskin13/stepper/examples"
)

func main() {
	db, err := examples.CreateMongoDatabase("stepepr")
	if err != nil {
		log.Fatal(err)
	}

	e := mongoEngine.NewMongoWithDb(db)
	s := stepper.NewService(e)

	s.RegisterJob(context.Background(), &stepper.JobConfig{
		Name:     "scheduled-job",
		Schedule: stepper.EverySecond().Interval(5),
	}, func(ctx stepper.Context) error {
		fmt.Println("Hello from scheduled job", time.Now())

		return nil
	})

	if err := s.Listen(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func createJob(s stepper.Stepper, customId string) {
	if err := s.CreateJob(context.Background(), &stepper.JobConfig{
		Name:     "job-with-custom-id",
		CustomId: customId,
		Pattern:  "@every 5s",
	}); err != nil {
		log.Fatal(err)
	}
}
