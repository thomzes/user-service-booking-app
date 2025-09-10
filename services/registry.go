package services

import (
	"github.com/thomzes/user-service-booking-app/repositories"
	services "github.com/thomzes/user-service-booking-app/services/user"
)

type Registry struct {
	repository repositories.IRepositoryRegistry
}

type IServiceRegistry interface {
	GetUser() services.IUserService
}

func NewServiceRegistry(repository repositories.IRepositoryRegistry) IServiceRegistry {
	return &Registry{repository: repository}
}

func (r *Registry) GetUser() services.IUserService {
	return services.NewUserService(r.repository)
}
