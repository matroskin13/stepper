package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/matroskin13/stepper"
	"github.com/matroskin13/stepper/engines/pg"
	"github.com/matroskin13/stepper/middlewares"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	/*db, err := mongo.NewMongo("mongodb://localhost:27017", "tasks")
	if err != nil {
		log.Fatal(err)
	}*/

	db, err := pg.NewPG("postgres://postgres:test@localhost:5432/postgres")
	if err != nil {
		log.Fatal(err)
	}

	s := stepper.NewService(db)

	prometheusMiddleware := middlewares.NewPrometheus()

	s.UseMiddleware(middlewares.LogMiddleware())
	s.UseMiddleware(prometheusMiddleware.GetMiddleware())
	s.UseMiddleware(middlewares.Retry(middlewares.RetryOptions{
		Interval:   time.Second * 5,
		MaxRetries: 3,
	}))

	s.TaskHandler("failed-task", func(ctx stepper.Context, data []byte) error {
		return fmt.Errorf("always return error")
	})

	if err := s.Publish(context.Background(), "failed-task", []byte("fail")); err != nil {
		log.Fatal(err)
	}

	go func() {
		http.ListenAndServe(":3999", promhttp.HandlerFor(prometheusMiddleware.GetRegistry(), promhttp.HandlerOpts{}))
	}()

	if err := s.Listen(context.Background()); err != nil {
		log.Fatal(err)
	}
}
