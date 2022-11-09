package middlewares

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/matroskin13/stepper"
)

type RetryOptions struct {
	MaxRetries int
	Interval   time.Duration
}

type retryState struct {
	Attempt int
}

func Retry(options RetryOptions) stepper.MiddlewareHandler {
	if options.Interval == 0 {
		options.Interval = time.Second * 10
	}

	return func(next stepper.MiddlewareFunc) stepper.MiddlewareFunc {
		return func(ctx stepper.Context, t *stepper.Task) error {
			var state retryState

			json.Unmarshal(t.MiddlewaresState["__retry"], &state)

			if err := next(ctx, t); err != nil {
				state.Attempt += 1

				newState, _ := json.Marshal(state)
				t.MiddlewaresState["__retry"] = newState

				if state.Attempt >= options.MaxRetries {
					ctx.SetRetryAfter(-1)
					return fmt.Errorf("a retry limit is exceeded")
				}

				ctx.SetRetryAfter(options.Interval)

				return err
			}

			return nil
		}
	}
}
