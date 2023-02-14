package main

import (
	"context"
	"fmt"
	"log"

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

	createJob(s, "first-id")
	createJob(s, "second-id")

	s.JobHandler("job-with-custom-id", func(ctx stepper.Context) error {
		customId := ctx.Task().CustomId

		fmt.Printf("[job-with-custom-id]: for task with id=%s \r\n", customId)

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
