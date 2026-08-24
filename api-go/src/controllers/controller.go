package controllers

import (
	"prueba-tecnica/api-go/src/service"

	"github.com/gofiber/fiber/v2"
)

type Controller struct {
	Service *service.Service
}

func New(serviceInstance *service.Service) *Controller {
	return &Controller{Service: serviceInstance}
}

func (controller *Controller) HealthHandler(contexto *fiber.Ctx) error {
	return controller.Service.Health(contexto)
}

func (controller *Controller) ManejarLogin(contexto *fiber.Ctx) error {
	return controller.Service.Login(contexto)
}

func (controller *Controller) ManejarFactorizacionQR(contexto *fiber.Ctx) error {
	return controller.Service.FactorizacionQR(contexto)
}
