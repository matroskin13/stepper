package examples

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateMongoDatabase(dbName string) (*mongo.Database, error) {
	cmdMonitor := &event.CommandMonitor{
		Started: func(_ context.Context, evt *event.CommandStartedEvent) {
			// log.Print(evt.Command)
		},
	}

	mongoHost := os.Getenv("MONGO_HOST")
	if mongoHost == "" {
		mongoHost = "mongodb://localhost:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoHost).SetMonitor(cmdMonitor))
	if err != nil {
		return nil, err
	}

	db := mongoClient.Database(dbName)

	return db, nil
}
