package main

import (
	"context"
	"fmt"
	"log"
	"strings"

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

	s.TaskHandler("task-with-threads", func(ctx stepper.Context, data []byte) error {
		fmt.Println("have received the word for splitting: ", string(data))

		for _, symbol := range strings.Split(string(data), "") {
			ctx.CreateSubtask(stepper.CreateTask{
				Data: []byte(symbol),
			})
		}

		return nil
	}).Subtask(func(ctx stepper.Context, data []byte) error {
		fmt.Printf("[letter-subtask]: have received symbol: %s\r\n", data)
		return nil
	}).OnFinish(func(ctx stepper.Context, data []byte) error {
		fmt.Println("subtasks are over")
		return nil
	})

	if err := s.Publish(context.Background(), "task-with-threads", []byte("hello")); err != nil {
		log.Fatal(err)
	}

	if err := s.Listen(context.Background()); err != nil {
		log.Fatal(err)
	}
}
