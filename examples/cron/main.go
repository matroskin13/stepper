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

	e := mongoEngine.NewMongo(db)
	s := stepper.NewService(e, e)

	s.RegisterJob(context.Background(), &stepper.JobConfig{
		Name:    "log-job",
		Pattern: "@every 15s",
	}, func(ctx stepper.Context) error {
		fmt.Println("wake up the log-job")

		ctx.CreateSubtask(stepper.CreateTask{
			Name: "log-subtask",
			Data: []byte("Hello 1 subtask"),
		})

		ctx.CreateSubtask(stepper.CreateTask{
			Name: "log-subtask",
			Data: []byte("Hello 2 subtask"),
		})

		return nil
	}).OnFinish(func(ctx stepper.Context, data []byte) error {
		fmt.Println("success job log-job")

		return nil
	})

	s.TaskHandler("log-subtask", func(ctx stepper.Context, data []byte) error {
		fmt.Println("message from subtask:", string(data))
		return nil
	})

	if err := s.Listen(context.Background()); err != nil {
		log.Fatal(err)
	}
}
