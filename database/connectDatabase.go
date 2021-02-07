package connectdatabase

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoInstance struct {
	Client *mongo.Client
	Db     *mongo.Database
}

// Mg is mongo Instance
var Mg mongoInstance

// Connect to mongo database
func Connect() error {
	const dbName = "thermotify"
	const mongoURI = "mongodb://localhost:27017/" + dbName

	// create connect mongo
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel() // close connection without module

	err = client.Connect(ctx) //connect database
	if err != nil {
		return err
	}

	db := client.Database(dbName)

	Mg = mongoInstance{
		Client: client,
		Db:     db,
	}
	return nil
}
