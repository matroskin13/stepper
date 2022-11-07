package stepper

type MiddlewareFunc func(ctx Context, t *Task) error
type MiddlewareHandler func(t MiddlewareFunc) MiddlewareFunc
