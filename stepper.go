package stepper

import "context"

type Stepper interface {
	TaskHandler(name string, handler Handler, middlewares ...MiddlewareHandler) HandlerStruct
	Listen(ctx context.Context) error
	Publish(ctx context.Context, name string, data []byte, options ...PublishOption) error
	RegisterJob(ctx context.Context, config *JobConfig, h JobHandler) HandlerStruct
	UseMiddleware(h MiddlewareHandler)
}
