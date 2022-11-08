package mongo

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/matroskin13/stepper"
	"github.com/matroskin13/stepper/tests"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMongo(t *testing.T) {
	tests.Run(t, func() stepper.Stepper {
		db, err := createTestMongoDatabase("tests")
		if err != nil {
			log.Fatal(err)
		}

		mongoEngine := NewMongo(db)

		return stepper.NewService(mongoEngine, mongoEngine)
	})
}

func createTestMongoDatabase(dbName string) (*mongo.Database, error) {
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
