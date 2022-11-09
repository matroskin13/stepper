package middlewares

import (
	"fmt"

	"github.com/matroskin13/stepper"
)

func LogMiddleware() stepper.MiddlewareHandler {
	return func(next stepper.MiddlewareFunc) stepper.MiddlewareFunc {
		return func(ctx stepper.Context, t *stepper.Task) error {
			fmt.Printf("take task=%s with body=%s\r\n", t.Name, string(t.Data))
			err := next(ctx, t)

			if err != nil {
				fmt.Printf("task task=%s has error: %s\r\n", t.Name, err)
			} else {
				fmt.Printf("complete task=%s\r\n", t.Name)
			}

			return err
		}
	}
}
