package middlewares

import (
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v2"
)

func JWT(secret []byte) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: secret,
	})
}
