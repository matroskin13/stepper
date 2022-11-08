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

    ctx := context.Background()

    service := stepper.NewService(mongoEngine)

    // Will publish a task on startup
    if err := service.Publish(ctx, "example-task", []byte("Hello world")); err != nil {
        log.Fatal(err)
    }

    s.TaskHandler("example-task", func(ctx stepper.Context, data []byte) error {
        fmt.Println(string(data))

        return nil
    })

    if err := s.Listen(ctx); err != nil {
        log.Fatal(err)
    }
}
```

## Table of Contents

- [Stepper](#stepper)
  - [Install](#install)
  - [Getting started](#getting-started)
  - [Table of Contents](#table-of-contents)
  - [Publish task](#publish-task)
    - [Simple way](#simple-way)
    - [Publish with delay](#publish-with-delay)

## Publish task

If you use the stepper you will use a lot of things but first of all you will publish and execute tasks. Let's discuss how you can publish tasks.

### Simple way

```go
service.Publish(context.Background(), "example-task", []byte("hello"))
```

The example shows the simple way to publish a task. The code will publish a task with a name **example-task** and content **hello**.

But also the stepper allows you to use additional options.

### Publish with delay

If you don't want to to execute a task immediately you can set up a delay.

```go
service.Publish(
    context.Background(),
    "example-task",
    []byte("hello"),
    stepper.SetDelay(time.Minute * 1),
)
```

Or you can use particular a date

```go
service.Publish(
    context.Background(),
    "example-task",
    []byte("hello"),
    stepper.LaunchAt(time.Now().Add(time.Minute * 10)),
)
```