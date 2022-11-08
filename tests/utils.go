package tests

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/matroskin13/stepper"
	"github.com/stretchr/testify/assert"
)

func getTestName(testFunc interface{}) string {
	fullName := runtime.FuncForPC(reflect.ValueOf(testFunc).Pointer()).Name()
	nameSliced := strings.Split(fullName, ".")

	return nameSliced[len(nameSliced)-1]
}

func listen(t *testing.T, ctx context.Context, s stepper.Stepper) {
	go func() {
		if err := s.Listen(ctx); err != nil {

		}
	}()
}

func readFor(ctx context.Context, s stepper.Stepper, task string) chan responseContainer {
	ch := make(chan responseContainer)

	go func() {
		s.TaskHandler(task, func(ctx stepper.Context, data []byte) error {
			ch <- responseContainer{ctx: ctx, data: data}
			return nil
		})

		if err := s.Listen(ctx); err != nil {
		}
	}()

	return ch
}

type responseContainer struct {
	ctx  stepper.Context
	data []byte
}

func waitChannelWithTimeout[T any](t *testing.T, ch chan (T), timeout time.Duration, msg string) (res T) {
	select {
	case <-time.After(timeout):
		t.Fatal("cannot wait channel", msg)
		return res
	case v := <-ch:
		return v
	}
}

func publishChannelWithTimeout[T any](t *testing.T, ch chan (T), v T, timeout time.Duration) {
	select {
	case <-time.After(timeout):
		t.Fatal("cannot publish channel")
	case ch <- v:
	}
}

func waitForValues(t *testing.T, reader chan responseContainer, values []string) {
	for _, item := range values {
		v := waitChannelWithTimeout(t, reader, time.Second*2, "")
		assert.Equal(t, item, string(v.data))
	}
}

type doorman struct {
	service  stepper.Stepper
	handlers map[string]stepper.Handler
	t        *testing.T
}

func newDoorman(t *testing.T, ctx context.Context, service stepper.Stepper) *doorman {
	go func() {
		if err := service.Listen(ctx); err != nil {
		}
	}()

	return &doorman{
		service:  service,
		handlers: make(map[string]stepper.Handler),
		t:        t,
	}
}

func (d *doorman) OnTask(name string, needFail bool, msg string) responseContainer {
	ch := make(chan responseContainer)

	if _, ok := d.handlers[name]; !ok {
		d.service.TaskHandler(name, func(ctx stepper.Context, data []byte) error {
			if err := d.handlers[name](ctx, data); err != nil {
				return err
			}

			return nil
		})
	}

	d.handlers[name] = func(ctx stepper.Context, data []byte) error {
		if needFail {
			ctx.SetRetryAfter(time.Second * 2)
			ch <- responseContainer{ctx: ctx, data: data}
			return fmt.Errorf("needFail=true")
		}

		ch <- responseContainer{ctx: ctx, data: data}
		return nil
	}

	return waitChannelWithTimeout(d.t, ch, time.Second*10, msg)
}

func (d *doorman) EqualValues(task string, values []string) {
	for _, v := range values {
		assert.Equal(d.t, v, string(d.OnTask(task, false, "").data))
	}
}
