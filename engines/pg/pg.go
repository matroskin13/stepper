package pg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matroskin13/stepper"

	sq "github.com/Masterminds/squirrel"
)

type _ctxKey string

var ctxKey = "__ctx:stepper:pg"

type PG struct {
	pool *pgxpool.Pool
}

func NewPG(host string) (*PG, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)

	conn, err := pgxpool.New(ctx, host)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, err
	}

	return &PG{
		pool: conn,
	}, nil
}

func (pg *PG) Init(ctx context.Context) error {
	if _, err := pg.pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS tasks (
		id TEXT,
		custom_id TEXT,
		name TEXT,
		data TEXT,
		job_id TEXT,
		parent TEXT,
		launch_at bigint,
		status TEXT,
		lock_at DATE,
		state TEXT,
		middlewares_state TEXT,
		error TEXT
	)`); err != nil {
		return err
	}

	if _, err := pg.pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS jobs (
		name TEXT UNIQUE,
		status TEXT,
		next_launch_at BIGINT,
		pattern TEXT
	)`); err != nil {
		return err
	}

	if _, err := pg.pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_full ON tasks(name, status, launch_at)`); err != nil {
		return err
	}

	if _, err := pg.pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_id ON tasks(id)`); err != nil {
		return err
	}

	return nil
}

func (pg *PG) Close() {
	pg.pool.Close()
}

func (pg *PG) GetRelatedTask(ctx context.Context, task *stepper.Task) (*stepper.Task, error) {
	return nil, nil
}

func (pg *PG) FindNextTask(ctx context.Context, statuses []string) (*stepper.Task, error) {
	tx, err := pg.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}

	var task Task

	if err := pgxscan.Get(
		ctx,
		tx,
		&task,
		"SELECT * FROM tasks WHERE launch_at <= $2 AND status = ANY($1) ORDER BY id LIMIT 1 FOR UPDATE SKIP LOCKED",
		statuses,
		time.Now().UnixNano(),
	); err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, tx.Commit(ctx)
		}

		tx.Rollback(ctx)

		return nil, fmt.Errorf("cannot find task: %w", err)
	}

	res := task.ToModel()

	res.EngineContext = context.WithValue(ctx, _ctxKey(ctxKey), tx)

	return res, nil
}

func (pg *PG) ReleaseTask(ctx context.Context, task *stepper.Task) error {
	tx, err := pg.getTx(task.EngineContext)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, "UPDATE tasks SET status = 'released' WHERE id = $1", task.ID); err != nil {
		tx.Rollback(ctx)
		return err
	}

	tx.Commit(ctx)

	return nil
}

func (pg *PG) WaitTaskForSubtasks(ctx context.Context, task *stepper.Task) error {
	return pg.runTX(task.EngineContext, func(tx pgx.Tx) error {
		_, err := tx.Exec(
			ctx,
			"UPDATE tasks SET status = 'waiting', launch_at = $1 WHERE id = $2",
			time.Now().Add(time.Second*1).UnixNano(),
			task.ID,
		)

		return err
	})
}

func (pg *PG) FailTask(ctx context.Context, task *stepper.Task, err error, timeout time.Duration) error {
	return pg.runTX(task.EngineContext, func(tx pgx.Tx) error {
		ms, _ := json.Marshal(task.MiddlewaresState)

		query := sq.
			Update("tasks").
			Set("status", "failed").
			Set("error", err.Error()).
			Set("middlewares_state", string(ms)).
			Where(sq.Eq{"id": task.ID}).
			PlaceholderFormat(sq.Dollar)

		if timeout != -1 {
			query = query.Set("launch_at", time.Now().Add(timeout).UnixNano())
		} else {
			query = query.Set("launch_at", nil)
		}

		sql, args, err := query.ToSql()
		if err != nil {
			return err
		}

		if _, err := tx.Exec(
			ctx,
			sql,
			args...,
		); err != nil {
			return err
		}

		return nil
	})
}

func (pg *PG) CreateTask(ctx context.Context, task *stepper.Task) error {
	ms, err := json.Marshal(task.MiddlewaresState)
	if err != nil {
		return err
	}

	if _, err := pg.pool.Exec(
		ctx,
		"INSERT INTO tasks VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
		task.ID,
		task.CustomId,
		task.Name,
		string(task.Data),
		task.JobId,
		task.Parent,
		task.LaunchAt.UnixNano(),
		task.Status,
		task.LockAt,
		string(task.State),
		string(ms),
	); err != nil {
		return err
	}

	return nil
}

func (pg *PG) GetUnreleasedTaskChildren(ctx context.Context, task *stepper.Task) (*stepper.Task, error) {
	var res *stepper.Task

	var t Task

	tx, err := pg.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}

	if err := pgxscan.Get(
		ctx,
		tx,
		&t,
		"SELECT * FROM tasks WHERE parent = $2 AND status = ANY($1) ORDER BY id LIMIT 1",
		[]string{"created", "in_progress"},
		task.ID,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, tx.Commit(ctx)
		}
		tx.Rollback(ctx)
		return nil, err
	}

	res = t.ToModel()

	return res, tx.Commit(ctx)
}

func (pg *PG) SetState(ctx context.Context, task *stepper.Task, state []byte) error {
	tx, err := pg.getTx(task.EngineContext)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, "UPDATE tasks SET state = $1 WHERE id = $2", string(state), task.ID); err != nil {
		return err
	}

	return nil
}

func (pg *PG) FindNextJob(ctx context.Context, statuses []string) (*stepper.Job, error) {
	tx, err := pg.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}

	var task Job

	if err := pgxscan.Get(
		ctx,
		tx,
		&task,
		"SELECT * FROM jobs WHERE next_launch_at <= $2 AND status = ANY($1)  LIMIT 1 FOR UPDATE SKIP LOCKED",
		statuses,
		time.Now().UnixNano(),
	); err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, tx.Commit(ctx)
		}

		tx.Rollback(ctx)

		return nil, fmt.Errorf("cannot find task: %w", err)
	}

	res := &stepper.Job{
		Name:         task.Name,
		Pattern:      task.Pattern,
		NextLaunchAt: time.Unix(0, task.NextLaunchAt),
		Status:       task.Status,
	}

	res.EngineContext = context.WithValue(ctx, _ctxKey(ctxKey), tx)

	return res, nil
}

func (pg *PG) GetUnreleasedJobChildren(ctx context.Context, jobName string) (*stepper.Task, error) {
	var res *stepper.Task

	var t Task

	tx, err := pg.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}

	if err := pgxscan.Get(
		ctx,
		tx,
		&t,
		"SELECT * FROM tasks WHERE job_id = $2 AND status = ANY($1) ORDER BY id LIMIT 1",
		[]string{"created", "in_progress", "released"},
		jobName,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			defer tx.Commit(ctx)

			//TODO hack for the locked transaction
			return &stepper.Task{
				Status: "in_progress",
			}, nil
		}
		tx.Rollback(ctx)
		return nil, err
	}

	defer tx.Commit(ctx)

	//TODO hack for the locked transaction
	if t.Status == "released" {
		return nil, nil
	}

	res = t.ToModel()

	return res, nil
}

func (pg *PG) Release(ctx context.Context, job *stepper.Job, nextLaunchAt time.Time) error {
	tx, err := pg.getTx(job.EngineContext)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, "UPDATE jobs SET status = 'released', next_launch_at = $2 WHERE name = $1", job.Name, nextLaunchAt.UnixNano()); err != nil {
		tx.Rollback(ctx)
		return err
	}

	tx.Commit(ctx)

	return nil
}

func (pg *PG) WaitForSubtasks(ctx context.Context, job *stepper.Job) error {
	return pg.runTX(job.EngineContext, func(tx pgx.Tx) error {
		_, err := tx.Exec(
			ctx,
			"UPDATE jobs SET status = 'waiting', next_launch_at = $1 WHERE name = $2",
			time.Now().Add(time.Second*1).UnixNano(),
			job.Name,
		)

		return err
	})
}

func (pg *PG) RegisterJob(ctx context.Context, cfg *stepper.JobConfig) error {
	nextLaunchAt, err := cfg.NextLaunch()
	if err != nil {
		return err
	}

	// TODO may be trouble with locking
	sql, values, err := sq.Insert("jobs").
		Columns("name", "status", "next_launch_at", "pattern").
		Values(cfg.Name, "created", nextLaunchAt.UnixNano(), cfg.Pattern).
		Suffix("ON CONFLICT (name) DO UPDATE SET next_launch_at = ?, pattern = ?", 1, cfg.Pattern).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	if _, err := pg.pool.Exec(ctx, sql, values...); err != nil {
		return err
	}

	return nil
}

func (pg *PG) getTx(ctx context.Context) (pgx.Tx, error) {
	tx, ok := ctx.Value(_ctxKey(ctxKey)).(pgx.Tx)
	if !ok {
		return nil, fmt.Errorf("cannot get pg context")
	}

	return tx, nil
}

func (pg *PG) runTX(ctx context.Context, callback func(tx pgx.Tx) error) error {
	tx, err := pg.getTx(ctx)
	if err != nil {
		return err
	}

	if err := callback(tx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}
