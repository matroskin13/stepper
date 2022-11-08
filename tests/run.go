package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/matroskin13/stepper"
	"github.com/rs/xid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

type ServiceCreator func() stepper.Stepper

type TestFunc func(t *testing.T, ctx context.Context, taskService stepper.Stepper)

func Run(t *testing.T, taskService ServiceCreator) {
	testCases := []TestFunc{
		simplePublish,
		publishAndRead,
		generateSubtasks,
		generateThreads,
		failTask,
	}

	for _, testCase := range testCases {
		t.Run(getTestName(testCase), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			testCase(t, ctx, taskService())
			cancel()
		})
	}
}

func simplePublish(t *testing.T, ctx context.Context, taskService stepper.Stepper) {
	name := xid.New().String()

	err := taskService.Publish(ctx, name, []byte("simple publish"))
	assert.Nil(t, err)
}

func publishAndRead(t *testing.T, ctx context.Context, taskService stepper.Stepper) {
	name := xid.New().String()

	for i := range lo.Range(3) {
		err := taskService.Publish(ctx, name, []byte(fmt.Sprintf("%v", i)))
		assert.Nil(t, err)
	}

	d := newDoorman(t, ctx, taskService)

	d.EqualValues(name, []string{"0", "1", "2"})
}

func generateSubtasks(t *testing.T, ctx context.Context, taskService stepper.Stepper) {
	name := xid.New().String()
	subtaskName := xid.New().String()

	finishAfterAllSubtasks := make(chan struct{})

	taskService.Publish(ctx, name, nil)
	taskService.TaskHandler(name, func(ctx stepper.Context, data []byte) error {
		for i := range lo.Range(3) {
			ctx.CreateSubtask(stepper.CreateTask{
				Name: subtaskName,
				Data: []byte(fmt.Sprintf("%v", i)),
			})
		}

		return nil
	}).OnFinish(func(ctx stepper.Context, data []byte) error {
		go func() {
			<-finishAfterAllSubtasks
		}()

		return nil
	})

	d := newDoorman(t, ctx, taskService)

	d.EqualValues(subtaskName, []string{"0", "1", "2"})

	publishChannelWithTimeout(t, finishAfterAllSubtasks, struct{}{}, time.Second*5)
}

func generateThreads(t *testing.T, ctx context.Context, taskService stepper.Stepper) {
	name := xid.New().String()

	subtasks := make(chan []byte, 3)
	finishAfterAllSubtasks := make(chan struct{})

	taskService.Publish(ctx, name, nil)
	taskService.TaskHandler(name, func(ctx stepper.Context, data []byte) error {
		for i := range lo.Range(3) {
			ctx.CreateSubtask(stepper.CreateTask{
				Data: []byte(fmt.Sprintf("%v", i)),
			})
		}

		return nil
	}).OnFinish(func(ctx stepper.Context, data []byte) error {
		go func() {
			<-finishAfterAllSubtasks
		}()

		return nil
	}).Subtask(func(ctx stepper.Context, data []byte) error {
		subtasks <- data

		return nil
	})

	listen(t, ctx, taskService)

	publishChannelWithTimeout(t, finishAfterAllSubtasks, struct{}{}, time.Second*5)
	assert.Len(t, subtasks, 3)

	assert.Equal(t, "0", string(<-subtasks))
	assert.Equal(t, "1", string(<-subtasks))
	assert.Equal(t, "2", string(<-subtasks))
}

func failTask(t *testing.T, ctx context.Context, taskService stepper.Stepper) {
	name := xid.New().String()

	taskService.Publish(ctx, name, []byte("failed"))

	d := newDoorman(t, ctx, taskService)

	d.OnTask(name, true, "wait for fail")
	startTime := time.Now()
	d.OnTask(name, false, "wait for failed message")
	assert.Equal(t, true, time.Now().After(startTime.Add(time.Second*2)), "failed message will receive without delay")
}
