package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.mongodb.org/mongo-driver/bson"

	connectdatabase "thermotify/database"
)

type hospital struct {
	HospitalID   string   `json:"hospitalId" bson:"hospitalId" `
	HospitalName string   `json:"hospitalName,omitempty" bson:"hospitalName"`
	SensorGroup  []string `json:"sensorGroup,omitempty" bson:"sensorGroup,omitempty"`
}

type group struct {
	GroupID    string   `json:"groupId" bson:"groupId"`
	GroupName  string   `json:"groupName" bson:"groupName"`
	LineToken  string   `json:"lineToken,omitempty" bson:"lineToken,omitempty"`
	SensorList []string `json:"sensorList,omitempty" bson:"sensorList,omitempty"`
}

type sensor struct {
	SensorID    string `json:"sensorId" bson:"sensorId"`
	SensorToken string `json:"sensorToken" bson:"sensorToken"`
	SensorName  string `json:"sensorName" bson:"sensorName"`
	MaxTemp     int    `json:"maxTemp" bson:"maxTemp"`
	MinTemp     int    `json:"minTemp" bson:"minTemp"`
	Notify      int    `json:"notify" bson:"notify"`
	AcceptData  int    `json:"acceptData" bson:"acceptData"`
}

type tempValues struct {
	SensorToken string `json:"sensorToken"`
	Time        uint32 `json:"time,omitempty"`
	Temp        int    `json:"temp"`
	Status      string `json:"status"`
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

	app.Get("/temp", getAllTemp)
	app.Get("/temp/:sensorToken", getTemp)
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
	hospitalID := c.Params("hospitalId")
	newGroup := new(group)
	if err := c.BodyParser(newGroup); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	hospitalsCollection := mg.Db.Collection("hospitals")

	// check hospitalId is exist
	isexist := bson.M{"hospitalId": hospitalID}
	if count, _ := hospitalsCollection.CountDocuments(c.Context(), isexist); count == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"err": "this hospitalID is not exist"})
	}

	// add new group to hospitals collection
	filter := bson.M{"hospitalId": hospitalID}
	_, err := hospitalsCollection.UpdateOne(c.Context(),
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
	groupsCollection := mg.Db.Collection("groups")
	_, err = groupsCollection.InsertOne(c.Context(), newGroup)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"err": "can not add this group into groups collection"})
	}
	return c.JSON(fiber.Map{"complete": "add this group successfully"})
}

func addSensor(c *fiber.Ctx) error {
	groupID := c.Params("groupId")
	newSensor := new(sensor)
	if err := c.BodyParser(newSensor); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	groupsCollection := mg.Db.Collection("groups")

	// check hospitalId is exist
	isexist := bson.M{"groupId": groupID}
	if count, _ := groupsCollection.CountDocuments(c.Context(), isexist); count == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"err": "this groupID is not exist"})
	}

	// add new sensor to groups collection
	filter := bson.M{"groupId": groupID}
	_, err := groupsCollection.UpdateOne(c.Context(),
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
	_, err = sensorsCollection.InsertOne(c.Context(), newSensor)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"err": "can not add this sensor into sensors collection"})
	}
	return c.JSON(fiber.Map{"complete": "add this sensor successfully"})
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

	temp.Time = uint32(time.Now().Unix())        // current time
	collection := mg.Db.Collection("tempValues") // choose collection
	insertResult, err := collection.InsertOne(c.Context(), temp)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	return c.JSON(fiber.Map{"insert temp complete": insertResult.InsertedID})
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
