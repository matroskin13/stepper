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

* [Stepper](#stepper)
  * [Publish task](#publish-task)
    * [Simple way](#simple-way)
    * [Publish with delay](#publish-with-delay)
  * [Execute a task](#execute-a-task)
    * [Simple way](#simple-way-1)
    * [Error handling](#error-handling)
    * [Bind a state](#bind-a-state)
  * [Subtasks](#subtasks)
    * [Create a subtask](#create-a-subtask)

## Publish task

If you use the stepper you will use a lot of things but first of all you will publish and execute tasks. Let's discuss how you can publish tasks.

### Simple way

```go
service.Publish(context.Background(), "example-task", []byte("hello"))
```

The example shows the simple way to publish a task. The code will publish a task with a name **example-task** and content **hello**.

But also the stepper allows you to use additional options.

### Publish with delay

If you don't want to execute a task immediately you can set up a delay.

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

## Execute a task

The second part of the Stepper is execution of tasks in queue.

### Simple way

```go
s.TaskHandler("example-task", func(ctx stepper.Context, data []byte) error {
    fmt.Println(string(data))

    return nil
})
```

The example shows the simple way to execute a task.

### Error handling

If your handler returns an error, a task will be returned to the queue. And the task will be held in the queue for 10 seconds. But you can set up a delay manually.

```go
s.TaskHandler("example-task", func(ctx stepper.Context, data []byte) error {
    ctx.SetRetryAfter(time.Minute) // will be returned in 1 minute
    return fmt.Errorf("some error")
})
```

### Bind a state

If you have a log running task, you can bind a state of task (cursor for example), and if your task failed you will be able to continue the task with the last state 

```go
s.TaskHandler("example-task", func(ctx stepper.Context, data []byte) error {
    var lastId int

    if err := ctx.BindState(&lastId); err != nil {
        return err
    }

    iter := getSomethingFromId(lastId) // something like a mongodb iterator or anything else

    for iter.Next() {
        lastId = ... // do something

        if err := ctx.SetState(lastId); err != nil {
            return err
        }
    }

    return nil
})
```

## Subtasks

The most powerful feature of the stepper is creating subtasks. The feature allows you to split a long-running task into separate tasks which will run on different nodes. And when all subtasks will be completed the stepper will call a `onFinish` hook of parent task.

### Create a subtask

The following example shows how to spawn subtasks within a main task.

```go
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
```

Or you can use existing a subtask:

```go
ctx.CreateSubtask(stepper.CreateTask{
    Name: "some-task",
    Data: []byte(symbol),
})
```