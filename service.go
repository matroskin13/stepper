package stepper

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/rs/xid"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

type Handler func(ctx Context, data []byte) error
type JobHandler func(ctx Context) error

type handlerStruct struct {
	handler          Handler
	onFinish         Handler
	onSubtask        Handler
	middlewares      []MiddlewareHandler
	jobHandler       JobHandler
	jobConfig        *JobConfig
	dependOnCustomId bool
}

func (h *handlerStruct) DependOnCustomId() HandlerStruct {
	h.dependOnCustomId = true

	return h
}

func (h *handlerStruct) OnFinish(handler Handler) HandlerStruct {
	h.onFinish = handler

	return h
}

func (h *handlerStruct) Subtask(handler Handler) HandlerStruct {
	h.onSubtask = handler

	return h
}

func (h *handlerStruct) UseMiddleware(middlewares ...MiddlewareHandler) {
	h.middlewares = middlewares
}

type HandlerStruct interface {
	OnFinish(h Handler) HandlerStruct
	Subtask(handler Handler) HandlerStruct
	UseMiddleware(middlewares ...MiddlewareHandler)
	DependOnCustomId() HandlerStruct
}

type Service struct {
	mongo     Engine
	jobEngine JobEngine

	jobs         map[string]*handlerStruct
	taskHandlers map[string]*handlerStruct

	middlewares []MiddlewareHandler
}

func NewService(engine Engine) Stepper {
	return &Service{
		jobs:         map[string]*handlerStruct{},
		taskHandlers: map[string]*handlerStruct{},
		mongo:        engine,
		jobEngine:    engine,
	}
}

func (s *Service) UseMiddleware(h MiddlewareHandler) {
	s.middlewares = append(s.middlewares, h)
}

func (s *Service) RegisterJob(ctx context.Context, config *JobConfig, h JobHandler) HandlerStruct {
	hs := handlerStruct{
		jobHandler: h,
		jobConfig:  config,
	}

	s.jobs[config.Name] = &hs

	return &hs
}

func (s *Service) TaskHandler(name string, h Handler) HandlerStruct {
	hs := handlerStruct{
		handler: h,
	}

	s.taskHandlers[name] = &hs

	return &hs
}

func (s *Service) createTask(ctx context.Context, task *CreateTask) error {
	launchAt := lo.Ternary(!task.LaunchAt.IsZero(), task.LaunchAt, time.Now())
	if task.LaunchAfter != 0 {
		launchAt = time.Now().Add(task.LaunchAfter)
	}

	return s.mongo.CreateTask(ctx, &Task{
		Name:             task.Name,
		Data:             task.Data,
		LaunchAt:         launchAt,
		Status:           "created",
		ID:               xid.New().String(),
		MiddlewaresState: map[string][]byte{},
		CustomId:         task.CustomId,
	})
}

func (s *Service) Publish(ctx context.Context, name string, data []byte, options ...PublishOption) error {
	created := &CreateTask{
		Name: name,
		Data: data,
	}

	for _, option := range options {
		option(created)
	}

	return s.createTask(ctx, created)
}

func (s *Service) Listen(ctx context.Context) error {
	if err := s.jobEngine.Init(ctx); err != nil {
		return err
	}

	for _, job := range s.jobs {
		if err := s.jobEngine.RegisterJob(ctx, job.jobConfig); err != nil {
			return fmt.Errorf("cannot register job=%s: %w", job.jobConfig.Name, err)
		}
	}

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return s.ListenJobs(gCtx)
	})

	g.Go(func() error {
		return s.ListenTasks(gCtx)
	})

	g.Go(func() error {
		return s.ListenWaitingJobs(ctx)
	})

	g.Go(func() error {
		return s.ListenWaitingTasks(ctx)
	})

	g.Go(func() error {
		return s.collectMetrics(ctx)
	})

	return g.Wait()
}

func (s *Service) collectMetrics(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second * 15):
			if err := s.mongo.CollectMetrics(ctx); err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (s *Service) handleTask(ctx context.Context, task *Task) error {
	var handler Handler
	var handlerMiddlewares []MiddlewareHandler

	if task.JobId == "" {
		name := task.Name
		isThread := strings.Contains(name, "__subtask:")

		if isThread {
			name = strings.TrimPrefix(name, "__subtask:")
		}

		_handler, ok := s.taskHandlers[name]
		if !ok {
			return nil
		}

		if _handler.dependOnCustomId {
			relatedTask, err := s.mongo.GetRelatedTask(ctx, task)
			if err != nil {
				return err
			}

			if relatedTask == nil {
				return nil
			}
		}

		handlerMiddlewares = _handler.middlewares

		handler = lo.Ternary(
			isThread && _handler.onSubtask != nil,
			_handler.onSubtask,
			_handler.handler,
		)
	} else {
		jobHandler, ok := s.jobs[task.JobId]
		if !ok {
			return nil
		}

		handlerMiddlewares = jobHandler.middlewares

		handler = func(ctx Context, data []byte) error {
			return jobHandler.jobHandler(ctx)
		}
	}

	_ctx := &taskContext{task: task, ctx: ctx, taskEngine: s.mongo}

	middlewares := lo.Flatten([][]MiddlewareHandler{s.middlewares, handlerMiddlewares})

	finalHandler := lo.Reduce(middlewares, func(r MiddlewareFunc, t MiddlewareHandler, i int) MiddlewareFunc {
		return t(r)
	}, func(ctx Context, task *Task) error {
		return handler(_ctx, task.Data)
	})

	if err := finalHandler(_ctx, task); err != nil {
		timeout := lo.Ternary(_ctx.retryAfter == 0, time.Second*10, _ctx.retryAfter)
		if err := s.mongo.FailTask(ctx, task, err, timeout); err != nil {
		}
		return nil
	}

	if len(_ctx.subtasks) > 0 {
		for _, subtask := range _ctx.subtasks {
			name := subtask.Name
			if name == "" {
				name = "__subtask:" + task.Name
			}

			launchAt := lo.Ternary(!subtask.LaunchAt.IsZero(), subtask.LaunchAt, time.Now())
			if subtask.LaunchAfter != 0 {
				launchAt = time.Now().Add(subtask.LaunchAfter)
			}

			if err := s.mongo.CreateTask(ctx, &Task{
				Name:             name,
				Parent:           task.ID,
				Status:           "created",
				LaunchAt:         launchAt,
				Data:             subtask.Data,
				ID:               xid.New().String(),
				MiddlewaresState: map[string][]byte{},
				CustomId:         subtask.CustomId,
			}); err != nil {
				return err
			}
		}

		if err := s.mongo.WaitTaskForSubtasks(ctx, task); err != nil {
			return fmt.Errorf("cannot set WaitTaskForSubtasks: %w", err)
		}
	} else {
		if err := s.mongo.ReleaseTask(ctx, task); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) ListenTasks(ctx context.Context) error {
	pool := Pool(ctx, runtime.NumCPU(), func(task *Task) {
		if err := s.handleTask(ctx, task); err != nil {
			fmt.Println(err)
		}
	})

	interval := time.Millisecond

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(interval):
			task, err := s.mongo.FindNextTask(ctx, []string{"created", "in_progress", "failed"})
			if err != nil {
				fmt.Println(err)
				continue
			}

			if task == nil {
				interval = time.Second
				continue
			}

			pool <- task

			interval = time.Millisecond
		}
	}
}

func (s *Service) handleWaitingTask(ctx context.Context, task *Task) error {
	subtask, err := s.mongo.GetUnreleasedTaskChildren(ctx, task)
	if err != nil {
		return fmt.Errorf("cannot get GetUnreleasedTaskChildren: %w", err)
	}

	if subtask == nil {
		hs, ok := s.taskHandlers[task.Name]
		if ok && task.JobId == "" && hs.onFinish != nil {
			if err := hs.onFinish(&taskContext{ctx: ctx, task: task}, task.Data); err != nil {
				// TODO need to fail the task
				return nil
			}
		}

		if err := s.mongo.ReleaseTask(ctx, task); err != nil {
			return fmt.Errorf("cannot release waiting task: %w", err)
		}
	} else {
		if err := s.mongo.WaitTaskForSubtasks(ctx, task); err != nil {
			return fmt.Errorf("cannot delay waiting subtask=%s: %w, %s, %s", task.ID, err, ctx.Err(), task.EngineContext.Err())
		}
	}

	return nil
}

func (s *Service) ListenWaitingTasks(ctx context.Context) error {
	interval := time.Millisecond

	pool := Pool(ctx, runtime.NumCPU(), func(task *Task) {
		if err := s.handleWaitingTask(ctx, task); err != nil {
			fmt.Println(err)
		}
	})

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(interval):
			task, err := s.mongo.FindNextTask(ctx, []string{"waiting"})
			if err != nil {
				continue
			}

			if task == nil {
				interval = time.Second
				continue
			}

			pool <- task

			interval = time.Millisecond

			continue
		}
	}
}

func (s *Service) ListenWaitingJobs(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second):
			job, err := s.jobEngine.FindNextJob(ctx, []string{"waiting"})
			if err != nil {
				fmt.Println(err)
				continue
			}

			if job == nil {
				continue
			}

			subtask, err := s.jobEngine.GetUnreleasedJobChildren(ctx, job.Name)
			if err != nil {
				return err
			}

			if subtask == nil {
				jobHs, ok := s.jobs[job.Name]
				if ok && jobHs.onFinish != nil {
					_ctx := &taskContext{ctx: ctx}
					if err := jobHs.onFinish(_ctx, nil); err != nil {
						continue
					}
				}

				job.CalculateNextLaunch()
				if err := s.jobEngine.Release(ctx, job, job.NextLaunchAt); err != nil {
					return err
				}
			} else {
				if err := s.jobEngine.WaitForSubtasks(ctx, job); err != nil {
					return err
				}
			}

			continue
		}
	}
}

func (s *Service) ListenJobs(ctx context.Context) error {
	interval := time.Millisecond

	for {
		select {
		case <-time.After(interval):
			job, err := s.jobEngine.FindNextJob(ctx, []string{"in_progress", "created", "released"})
			if err != nil {
				fmt.Println(err)
				continue
			}

			if job == nil {
				interval = time.Second
				continue
			}

			if err := s.mongo.CreateTask(ctx, &Task{
				ID:               xid.New().String(),
				Name:             "__job:" + job.Name,
				JobId:            job.Name,
				Status:           "created",
				LaunchAt:         time.Now(),
				Data:             nil,
				MiddlewaresState: map[string][]byte{},
				CustomId:         "",
			}); err != nil {
				return err
			}

			if err := s.jobEngine.WaitForSubtasks(ctx, job); err != nil {
				return err
			}

			interval = time.Millisecond
		case <-ctx.Done():
			return nil
		}
	}
}
