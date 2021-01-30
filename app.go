package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
		if err != nil {
			panic(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		err = client.Connect(ctx)
		if err != nil {
			panic(err)
		}
		defer cancel()
		fmt.Println("asdasd")
		return c.SendString(fmt.Sprintln("connect :", cancel))
	})

	app.Listen(":80")
}
