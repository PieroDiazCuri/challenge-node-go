package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"prueba-tecnica/api-go/src/controllers"
	"prueba-tecnica/api-go/src/middlewares"
	"prueba-tecnica/api-go/src/routes"
	"prueba-tecnica/api-go/src/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

var jwtSecret = []byte("your-secret-key")

func main() {
	app := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE",
	}))

	service := service.New(jwtSecret)
	controller := controllers.New(service)
	jwtMiddleware := middlewares.JWT(jwtSecret)

	routes.Routes(app, jwtMiddleware, controller.HealthHandler, controller.ManejarFactorizacionQR, controller.ManejarLogin)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))

	apiKey := os.Getenv("API_KEY")
	fmt.Println("API_KEY desde Azure:", apiKey)
}
