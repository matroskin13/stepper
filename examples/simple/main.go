package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/matroskin13/stepper"
	mongoEngine "github.com/matroskin13/stepper/engines/mongo"
	"github.com/matroskin13/stepper/examples"
	"github.com/matroskin13/stepper/middlewares"
)

type State struct {
	Count int
}

func main() {
	db, err := examples.CreateMongoDatabase("stepepr")
	if err != nil {
		log.Fatal(err)
	}

	e := mongoEngine.NewMongo(db)
	s := stepper.NewService(e, e)

	s.UseMiddleware(middlewares.LogMiddleware())

	s.TaskHandler("simple", func(ctx stepper.Context, data []byte) error {
		fmt.Println(string(data))

		return nil
	})

	for i := 0; i < 10; i++ {
		if err := s.Publish(context.Background(), "simple", []byte(fmt.Sprintf("hello from %v", i)), stepper.LaunchAt(time.Now().Add(time.Second*10))); err != nil {
			log.Fatal(err)
		}
	}

	if err := s.Listen(context.Background()); err != nil {
		log.Fatal(err)
	}
}
