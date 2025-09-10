package controllers

import (
	controllers "github.com/thomzes/user-service-booking-app/controllers/user"
	"github.com/thomzes/user-service-booking-app/services"
)

type Registry struct {
	service services.IServiceRegistry
}

type IControllerRegistry interface {
	GetUserController() controllers.IUserController
}

func NewControllerRegistry(service services.IServiceRegistry) IControllerRegistry {
	return &Registry{service: service}
}

func (u *Registry) GetUserController() controllers.IUserController {
	return controllers.NewUserController(u.service)
}
