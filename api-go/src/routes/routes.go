package routes

import "github.com/gofiber/fiber/v2"

func Routes(app *fiber.App, jwtMiddleware fiber.Handler, healthHandler fiber.Handler, qrHandler fiber.Handler, loginHandler fiber.Handler) {
	apiGroup := app.Group("/api")

	apiGroup.Get("/health", healthHandler)
	apiGroup.Post("/qr-factorization", jwtMiddleware, qrHandler)
	apiGroup.Post("/login", loginHandler)
}
