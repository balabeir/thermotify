package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	connectdatabase "thermotify/database"
)

type hospital struct {
	HospitalID   string   `json:"hospitalID,omitempty" bson:"_id"`
	HospitalName string   `json:"hospitalName" bson:"hospitalName"`
	SensorGroup  []string `json:"sensorGroup,omitempty" bson:"sensorGroup,omitempty"`
}

type group struct {
	GroupID    string   `json:"groupID,omitempty" bson:"_id"`
	GroupName  string   `json:"groupName" bson:"groupName"`
	LineToken  string   `json:"lineToken,omitempty" bson:"lineToken,omitempty"`
	SensorList []string `json:"sensorList,omitempty" bson:"sensorList,omitempty"`
}

type sensor struct {
	SensorID      string  `json:"sensorID,omitempty" bson:"_id"`
	SensorToken   string  `json:"sensorToken" bson:"sensorToken"`
	SensorName    string  `json:"sensorName" bson:"sensorName"`
	MaxTemp       float64 `json:"maxTemp" bson:"maxTemp"`
	MinTemp       float64 `json:"minTemp" bson:"minTemp"`
	Notify        int     `json:"notify" bson:"notify"`
	AcceptData    int     `json:"acceptData" bson:"acceptData"`
	CurrentStatus string  `json:"currentStatus,omitempty" bson:"currentStatus"`
}

type tempValues struct {
	SensorToken string  `json:"sensorToken"`
	Time        uint32  `json:"time,omitempty"`
	Temp        float64 `json:"temp"`
	Status      string  `json:"status"`
}

var mg = connectdatabase.Mg

func main() {
	if err := connectdatabase.Connect(); err != nil {
		log.Fatal(err)
	}
	mg = connectdatabase.Mg

	app := fiber.New()
	app.Use(logger.New())

	// app.Get("/hospitals", getHospitals)
	app.Post("/hospital", newHospital)
	// app.Put("/hospital", changeHospitalName)

	app.Post("/group/:hospitalId", addGroup)
	app.Post("/sensor/:groupId", addSensor)

	// app.Get("/temp", getAllTemp)
	// app.Get("/temp/:sensorToken", getTemp)
	app.Post("/temp", addTemp)

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
	// assume _id is hospitalID
	newHospital.HospitalID = primitive.NewObjectID().Hex()

	// insert new hospital
	_, err := hospitalCollection.InsertOne(context.Background(), newHospital)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	return c.JSON(fiber.Map{"complete": newHospital})
}

func changeHospitalName(c *fiber.Ctx) error {
	hospitalCollection := mg.Db.Collection("hospitals")

	changeName := new(hospital)

	if err := c.BodyParser(changeName); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	updateResult, err := hospitalCollection.UpdateOne(
		context.Background(),
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
	hospitalID := c.Params("hospitalId")
	newGroup := new(group)
	if err := c.BodyParser(newGroup); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}
	hospitalsCollection := mg.Db.Collection("hospitals")

	// check hospitalId is exist
	existFilter := bson.M{"_id": hospitalID}
	if err := fieldIsExist(hospitalsCollection.Name(), existFilter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"err": "this hospitalID is not exist"})
	}

	newGroup.GroupID = primitive.NewObjectID().Hex()

	// add new group to hospitals collection
	groupsCollection := mg.Db.Collection("groups")
	filter := bson.M{"_id": hospitalID}
	_, err := hospitalsCollection.UpdateOne(
		context.Background(),
		filter,
		bson.M{
			"$push": bson.M{"sensorGroup": newGroup.GroupID},
		},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"err": "can not add this group into hospital"})
	}

	// add new group to groups collection
	_, err = groupsCollection.InsertOne(context.Background(), newGroup)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"err": "can not add this group into groups collection"})
	}
	return c.JSON(fiber.Map{"complete": newGroup})
}

func addSensor(c *fiber.Ctx) error {
	groupID := c.Params("groupId")
	newSensor := new(sensor)
	if err := c.BodyParser(newSensor); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	groupsCollection := mg.Db.Collection("groups")

	// check hospitalId is exist
	existFilter := bson.M{"_id": groupID}
	if err := fieldIsExist(groupsCollection.Name(), existFilter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"err": "this hospitalID is not exist"})
	}

	newSensor.SensorID = primitive.NewObjectID().Hex()
	newSensor.CurrentStatus = "init"

	// add new sensor to groups collection
	filter := bson.M{"_id": groupID}
	_, err := groupsCollection.UpdateOne(context.Background(),
		filter,
		bson.M{
			"$push": bson.M{"sensorList": newSensor.SensorID},
		},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"err": "can not add this sensor into groups"})
	}

	// add new sensor to sensors collection
	sensorsCollection := mg.Db.Collection("sensors")
	_, err = sensorsCollection.InsertOne(context.Background(), newSensor)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"err": "can not add this sensor into sensors collection"})
	}
	return c.JSON(fiber.Map{"complete": newSensor})
}

func getTemp(c *fiber.Ctx) error {
	SensorID := c.Params("sensorId")
	filter := bson.M{"sensorId": SensorID}
	tempData := find("tempValues", filter, c)

	return tempData
}

func getAllTemp(c *fiber.Ctx) error {
	tempData := find("tempValues", bson.M{}, c)
	return tempData
}

func addTemp(c *fiber.Ctx) error {
	temp := new(tempValues)
	// parse json to struct
	if err := c.BodyParser(&temp); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	temp.Status = checkStatus(temp.SensorToken, temp.Temp) // set status
	temp.Time = uint32(time.Now().Unix())                  // current time

	// insert temp to tempValues collection
	tempCollection := mg.Db.Collection("tempValues") // choose collection
	_, err := tempCollection.InsertOne(context.Background(), temp)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"err": "can not insert temp into tempValues collection"})
	}

	return c.JSON(fiber.Map{"complete": temp})
}

// Sub function

func fieldIsExist(collectionName string, filter bson.M) error {
	collection := mg.Db.Collection(collectionName)
	if count, err := collection.CountDocuments(context.Background(), filter); count == 0 {
		return err
	}
	return nil
}

func checkStatus(sensorToken string, temp float64) (status string) {
	var threshold struct {
		Min           float64 `bson:"minTemp"`
		Max           float64 `bson:"maxTemp"`
		CurrentStatus string  `bson:"currentStatus"`
	}

	collection := mg.Db.Collection("sensors")
	filter := bson.M{"sensorToken": sensorToken}

	if err := collection.FindOne(context.TODO(), filter).Decode(&threshold); err != nil {
		log.Fatal("err ", err)
	}

	// check temp are not out of threshold
	if temp >= threshold.Max || temp <= threshold.Min {
		status = "bad"
	} else {
		status = "good"
	}

	// check for current status must be change
	if threshold.CurrentStatus != status {
		_, err := collection.UpdateOne(context.TODO(),
			filter,
			bson.M{"$set": bson.M{"currentStatus": status}})
		if err != nil {
			log.Fatal("err ", err)
		}
	}

	// return status
	return
}

func find(collection string, filter bson.M, c *fiber.Ctx) error {
	dbCollection := mg.Db.Collection(collection) // choose collection

	// filter := bson.M{"hospitalId": "111111"}
	findFilter := filter

	cursor, err := dbCollection.Find(context.Background(), findFilter)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}
	defer cursor.Close(context.Background())

	var data []bson.M
	if err = cursor.All(context.Background(), &data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}
	return c.JSON(data)
}
