package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/matroskin13/stepper"
	"github.com/matroskin13/stepper/examples"
	"github.com/matroskin13/stepper/middlewares"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	mongoEngine "github.com/matroskin13/stepper/engines/mongo"
)

func main() {
	db, err := examples.CreateMongoDatabase("stepepr")
	if err != nil {
		log.Fatal(err)
	}

	e := mongoEngine.NewMongoWithDb(db)
	s := stepper.NewService(e)

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
		prometheus.DefaultRegisterer.MustRegister(prometheusMiddleware.GetRegistry())
		http.ListenAndServe(":3999", promhttp.Handler())
	}()

	if err := s.Listen(context.Background()); err != nil {
		log.Fatal(err)
	}
}
