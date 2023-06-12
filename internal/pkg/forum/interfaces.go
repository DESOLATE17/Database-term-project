package forum

import "github.com/DESOLATE17/Database-term-project/internal/models"

type UseCase interface {
	GetUser(user models.User) (models.User, error)
	CreateUser(user models.User) ([]models.User, error)
}

type Repository interface {
	GetUser(name string) (models.User, error)
	CheckUserEmailUniq(usersS models.User) ([]models.User, error)
	CreateUser(user models.User) (models.User, error)
}
