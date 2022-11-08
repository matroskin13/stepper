# Stepper

A simple, efficient, concurrent task runner.

* **Simple.** Run tasks and schedule jobs with GO.
* **Database agnostic.** Stepper supports MongoDB, Postgresql.
* **Concurrent.** Stepper can be used in an unlimited number of instances.
* **Scalable.** Split one task into small subtasks which will run on different nodes.

## Install

```bash
go get github.com/matroskin13/stepper
```

## Getting started

```go
package main

import (
    "log"

    "github.com/matroskin13/stepper"
    "github.com/matroskin13/stepper/engines/mongo"
)

func main() {
    mongoEngine, err := mongo.NewMongo("mongodb://localhost:27017", "example_database")
    if err != nil {
        log.Fatal(err)
    }

    service := stepper.NewService(mongoEngine)

    // Will publish a task on startup
    if err := service.Publish(context.Background(), "example-task", []byte("Hello world")); err != nil {
        log.Fatal(err)
    }

    s.TaskHandler("example-task", func(ctx stepper.Context, data []byte) error {
        fmt.Println(string(data))

        return nil
    })

    if err := s.Listen(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```