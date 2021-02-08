package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.mongodb.org/mongo-driver/bson"

	connectdatabase "thermotify/database"
)

type hospital struct {
	ID           string  `json:"_id,omitempty" bson:"_id,omitempty"`
	HospitalID   string  `json:"hospitalId" bson:"hospitalId" `
	HospitalName string  `json:"hospitalName,omitempty" bson:"hospitalName"`
	GroupSensor  []group `json:"groupSensor,omitempty" bson:"groupSensor,omitempty"`
}

type group struct {
	GroupID    string   `json:"groupId" bson:"groupId"`
	GroupName  string   `json:"groupName" bson:"groupName"`
	LineToken  string   `json:"lineToken,omitempty" bson:"lineToken"`
	SensorList []sensor `json:"sensorList,omitempty" bson:"sensorList"`
}

type sensor struct {
	SensorToken string `json:"sensorToken" bson:"sensorToken"`
	SensorName  string `json:"sensorName" bson:"sensorName"`
	MaxTemp     int    `json:"maxTemp" bson:"maxTemp"`
	MinTemp     int    `json:"minTemp" bson:"minTemp"`
	Notify      int    `json:"notify" bson:"notify"`
	AcceptData  int    `json:"acceptData" bson:"acceptData"`
}

var mg = connectdatabase.Mg

func main() {
	if err := connectdatabase.Connect(); err != nil {
		log.Fatal(err)
	}
	mg = connectdatabase.Mg

	app := fiber.New()
	app.Use(logger.New())

	app.Get("/hospitals", getHospitals)
	app.Post("/hospital", newHospital)
	app.Put("/hospital", changeHospitalName)

	app.Put("/group/:hospitalId", addGroup)

	app.Get("/temp", getAllTemp)
	app.Get("/temp/:sensorId", getSensorTemp)

	app.Listen(":80")
}

func getHospitals(c *fiber.Ctx) error {
	hospitalData := find("hospitals", bson.M{}, c)

	return hospitalData
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

	return c.JSON(fiber.Map{"Insert successfully": insertResult.InsertedID})
}

func changeHospitalName(c *fiber.Ctx) error {
	hospitalCollection := mg.Db.Collection("hospitals")

	changeName := new(hospital)

	if err := c.BodyParser(changeName); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

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

func addGroup(c *fiber.Ctx) error {
	collection := mg.Db.Collection("hospitals")

	addGroup := new(group)
	if err := c.BodyParser(addGroup); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	filter := bson.M{"hospitalId": c.Params("hospitalId")}

	updateResult, err := collection.UpdateOne(c.Context(),
		filter,
		bson.M{
			"$push": bson.M{"groupSensor": addGroup},
		},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}
	return c.JSON(fiber.Map{"Matched Document": updateResult.MatchedCount})
}

func getSensorTemp(c *fiber.Ctx) error {
	SensorID := c.Params("sensorId")
	filter := bson.M{"sensorId": SensorID}
	tempData := find("tempValue", filter, c)

	return tempData
}

func getAllTemp(c *fiber.Ctx) error {
	tempData := find("tempValue", bson.M{}, c)
	return tempData
}

func find(collection string, filter bson.M, c *fiber.Ctx) error {
	dbCollection := mg.Db.Collection(collection) // choose collection

	// filter := bson.M{"hospitalId": "111111"}
	findFilter := filter

	cursor, err := dbCollection.Find(c.Context(), findFilter)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}
	var data []bson.M
	if err = cursor.All(c.Context(), &data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}
	return c.JSON(data)
}
