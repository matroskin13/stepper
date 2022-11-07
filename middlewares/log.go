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
			fmt.Printf("result for task=%s: %v\r\n", t.Name, err)

			return err
		}
	}
}
