package mongo

import (
	"context"
	"time"

	"github.com/matroskin13/stepper"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	jobs  *mongo.Collection
	tasks *mongo.Collection
}

func NewMongoWithDb(db *mongo.Database) *Mongo {
	return &Mongo{
		jobs:  db.Collection("jobs"),
		tasks: db.Collection("tasks"),
	}
}

func NewMongo(host string, database string) (*Mongo, error) {
	db, err := createMongoDatabase(host, database)
	if err != nil {
		return nil, err
	}

	return &Mongo{
		jobs:  db.Collection("jobs"),
		tasks: db.Collection("tasks"),
	}, nil
}

func (m *Mongo) RegisterJob(ctx context.Context, cfg *stepper.JobConfig) error {
	nextLaunchAt, err := cfg.NextLaunch()
	if err != nil {
		return err
	}

	query := bson.M{"name": cfg.Name}
	update := bson.M{
		"nextLaunchAt": nextLaunchAt,
		"name":         cfg.Name,
		"tags":         cfg.Tags,
		"pattern":      cfg.Pattern,
		"status":       "created",
	}

	opts := options.FindOneAndReplace().SetUpsert(true).SetReturnDocument(options.After)

	return m.jobs.FindOneAndReplace(ctx, query, update, opts).Err()
}

func (m *Mongo) CreateTask(ctx context.Context, task *stepper.Task) error {
	t := Task{}
	t.FromModel(task)
	_, err := m.tasks.InsertOne(ctx, t)
	return err
}

func (m *Mongo) SetState(ctx context.Context, id string, state []byte) error {
	query := bson.M{"id": id}
	update := bson.M{"$set": bson.M{"state": state}}

	if err := m.tasks.FindOneAndUpdate(ctx, query, update).Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}

		return err
	}

	return nil
}

func (m *Mongo) FindNextTask(ctx context.Context, statuses []string) (*stepper.Task, error) {
	var job Task

	query := bson.M{
		"status": bson.M{"$in": statuses},
		"launchAt": bson.M{
			"$lte": time.Now(),
		},
		"$or": []bson.M{
			{"lock_at": nil},
			{"lock_at": bson.M{"$lte": time.Now().Add(time.Minute * -1)}}, // TODO pass right timeout
		},
	}

	update := bson.M{
		"$set": bson.M{
			"lock_at": time.Now(),
			"status":  "in_progress",
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	if err := m.tasks.FindOneAndUpdate(ctx, query, update, opts).Decode(&job); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}

		return nil, err
	}

	return job.ToModel(), nil
}

func (m *Mongo) FindNextJob(ctx context.Context, statuses []string) (*stepper.Job, error) {
	var _job job

	query := bson.M{
		"status": bson.M{"$in": statuses},
		"nextLaunchAt": bson.M{
			"$lte": time.Now(),
		},
		"$or": []bson.M{
			{"lock_at": nil},
			{"lock_at": bson.M{"$lte": time.Now().Add(time.Minute * 5)}}, // TODO pass right timeout
		},
	}

	update := bson.M{
		"$set": bson.M{
			"lock_at": time.Now(),
			"status":  "in_progress",
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	if err := m.jobs.FindOneAndUpdate(ctx, query, update, opts).Decode(&_job); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}

		return nil, err
	}

	return _job.ToModel(), nil
}

func (m *Mongo) GetUnreleasedJobChildren(ctx context.Context, jobId string) (*stepper.Task, error) {
	var task Task

	query := bson.M{
		"status": bson.M{"$in": []string{"created", "in_progress"}},
		"jobId":  jobId,
	}

	if err := m.tasks.FindOne(ctx, query).Decode(&task); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}

		return nil, err
	}

	return task.ToModel(), nil
}

func (m *Mongo) GetUnreleasedTaskChildren(ctx context.Context, id string) (*stepper.Task, error) {
	var task Task

	query := bson.M{
		"status": bson.M{"$in": []string{"created", "in_progress"}},
		"parent": id,
	}

	if err := m.tasks.FindOne(ctx, query).Decode(&task); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}

		return nil, err
	}

	return task.ToModel(), nil
}

func (m *Mongo) GetRelatedTask(ctx context.Context, task string, id string) (*stepper.Task, error) {
	query := bson.M{"custom_id": id, "name": task, "status": bson.M{"$ne": "released"}}

	var e Task

	if err := m.tasks.FindOne(ctx, query).Decode(&e); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}

		return nil, err
	}

	return e.ToModel(), nil
}

func (m *Mongo) Release(ctx context.Context, name string, nextTimeLaunch time.Time) error {
	return m.jobs.FindOneAndUpdate(
		ctx,
		bson.M{"name": name},
		bson.M{"$set": bson.M{
			"lock_at":      nil,
			"status":       "released",
			"nextLaunchAt": nextTimeLaunch,
		}},
	).Err()
}

func (m *Mongo) FailTask(ctx context.Context, id string, handlerErr error, timeout time.Duration) error {
	return m.tasks.FindOneAndUpdate(
		ctx,
		bson.M{"id": id},
		bson.M{"$set": bson.M{
			"launchAt": time.Now().Add(timeout),
			"lock_at":  nil,
			"status":   "failed",
			"error":    handlerErr.Error(),
		}},
	).Err()
}

func (m *Mongo) ReleaseTask(ctx context.Context, id string) error {
	return m.tasks.FindOneAndUpdate(
		ctx,
		bson.M{"id": id},
		bson.M{"$set": bson.M{
			"lock_at": nil,
			"status":  "released",
		}},
	).Err()
}

func (m *Mongo) WaitForSubtasks(ctx context.Context, name string) error {
	return m.jobs.FindOneAndUpdate(
		ctx,
		bson.M{"name": name},
		bson.M{"$set": bson.M{
			"lock_at":      nil,
			"status":       "waiting",
			"nextLaunchAt": time.Now().Add(time.Second * 5),
		}},
	).Err()
}

func (m *Mongo) WaitTaskForSubtasks(ctx context.Context, id string) error {
	return m.tasks.FindOneAndUpdate(
		ctx,
		bson.M{"id": id},
		bson.M{"$set": bson.M{
			"lock_at":  nil,
			"status":   "waiting",
			"launchAt": time.Now().Add(time.Second * 1),
		}},
	).Err()
}
