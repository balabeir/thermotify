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

func main() {

	app := fiber.New()

	app.Use(logger.New())

	app.Get("/", func(c *fiber.Ctx) error {
		// create connect mongo
		client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
		if err != nil {
			log.Fatal(err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		err = client.Connect(ctx) //connect database
		if err != nil {
			log.Fatal(err)
		}
		defer cancel() // close connection without module

		db := client.Database("thermotify")              // use database
		hospitalCollection := db.Collection("hospitals") // choose collection

		filter := bson.M{"hospitalId": "111111"}

		cursor, err := hospitalCollection.Find(ctx, filter)
		if err != nil {
			log.Fatal(err)
		}
		defer cursor.Close(ctx) // close cursor while not use

		var hospitalData bson.Raw
		// find all document
		for cursor.Next(ctx) {
			var result hospital
			if err = cursor.Decode(&result); err != nil {
				log.Fatal(err)
			}
			hospitalData = cursor.Current

		}
		return c.JSON(hospitalData)
	})

	app.Listen(":80")
}
