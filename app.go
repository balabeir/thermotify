package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type hospital struct {
	_id          string
	hospitalID   string
	hospitalName string
	groupSensor  []group
}

type group struct {
	groupID    string
	groupName  string
	lineToken  string
	sensorList []sensor
}

type sensor struct {
	sensorID    string
	sensorToken string
	maxTemp     int
	minTemp     int
	notify      int
	acceptData  int
}

type mongoInstance struct {
	Client *mongo.Client
	Db     *mongo.Database
}

var mg mongoInstance

const dbName = "thermotify"
const mongoURI = "mongodb://localhost:27017/" + dbName

func connect() error {
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

	mg = mongoInstance{
		Client: client,
		Db:     db,
	}
	return nil
}

func main() {
	if err := connect(); err != nil {
		log.Fatal(err)
	}

	app := fiber.New()

	app.Use(logger.New())

	app.Get("/", func(c *fiber.Ctx) error {

		hospitalCollection := mg.Db.Collection("hospitals") // choose collection

		// filter := bson.M{"hospitalId": "111111"}
		filter := bson.M{}

		cursor, err := hospitalCollection.Find(c.Context(), filter)
		if err != nil {
			log.Fatal(err)
		}
		defer cursor.Close(c.Context()) // close cursor while not use

		var hospitalData []bson.M
		if err = cursor.All(c.Context(), &hospitalData); err != nil {
			log.Fatal(err)
		}

		return c.JSON(hospitalData)
	})

	app.Listen(":80")
}
