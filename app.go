package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type hospital struct {
	ID           string  `json:"_id,omitempty" bson:"_id,omitempty"`
	HospitalID   string  `json:"hospitalId" bson:"hospitalId" `
	HospitalName string  `json:"hospitalName" bson:"hospitalName"`
	GroupSensor  []group `json:"groupSensor,omitempty" bson:"groupSensor,omitempty"`
}

type group struct {
	GroupID    string   `json:"groupId" bson:"groupId"`
	GroupName  string   `json:"groupName" bson:"groupName"`
	LineToken  string   `json:"lineToken" bson:"lineToken"`
	SensorList []sensor `json:"sensorList" bson:"sensorList"`
}

type sensor struct {
	SensorID    string `json:"sensorId" bson:"sensorId"`
	SensorToken string `json:"sensorToken" bson:"sensorToken"`
	MaxTemp     int    `json:"maxTemp" bson:"maxTemp"`
	MinTemp     int    `json:"minTemp" bson:"minTemp"`
	Notify      int    `json:"notify" bson:"notify"`
	AcceptData  int    `json:"acceptData" bson:"acceptData"`
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

	app.Get("/hospitals", getHospitals)
	app.Post("/hospital", newHospital)
	app.Put("/hospital", changeHospitalName)

	app.Listen(":80")
}

func getHospitals(c *fiber.Ctx) error {

	hospitalCollection := mg.Db.Collection("hospitals") // choose collection

	// filter := bson.M{"hospitalId": "111111"}
	filter := bson.M{}

	cursor, err := hospitalCollection.Find(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}
	defer cursor.Close(c.Context()) // close cursor while not use

	var hospitalData []bson.M
	if err = cursor.All(c.Context(), &hospitalData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	return c.JSON(hospitalData)
}

func newHospital(c *fiber.Ctx) error {

	hospitalCollection := mg.Db.Collection("hospitals")

	newHospital := new(hospital)
	if err := c.BodyParser(newHospital); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	newHospital.ID = "" // force MongoDB to always set its own generated ObjectIDs
	// insert new hospital
	insertResult, err := hospitalCollection.InsertOne(c.Context(), newHospital)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	// get the record has just been inserted
	// var createdHospital bson.M
	// filter := bson.D{{Key: "_id", Value: insertResult.InsertedID}}
	// if err := hospitalCollection.FindOne(c.Context(), filter).Decode(&createdHospital); err != nil {
	// 	return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	// }

	return c.JSON(fiber.Map{"Insert successfully": insertResult.InsertedID})
}

func changeHospitalName(c *fiber.Ctx) error {
	hospitalCollection := mg.Db.Collection("hospitals")

	changeName := new(hospital)

	if err := c.BodyParser(changeName); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	fmt.Println("hospitalID =", changeName.HospitalID)

	updateResult, err := hospitalCollection.UpdateOne(
		c.Context(),
		bson.M{"hospitalId": changeName.HospitalID},
		bson.M{
			"$set": bson.M{"hospitalName": changeName.HospitalName},
		},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	return c.JSON(fiber.Map{"update hospital name complete": updateResult})
}
