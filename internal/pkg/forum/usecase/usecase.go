package usecase

import (
	"github.com/DESOLATE17/Database-term-project/internal/models"
	"github.com/DESOLATE17/Database-term-project/internal/pkg/forum"
)

type UseCase struct {
	repo forum.Repository
}

func NewRepoUsecase(repo forum.Repository) forum.UseCase {
	return &UseCase{repo: repo}
}

func (u *UseCase) GetUser(user models.User) (models.User, error) {
	return u.repo.GetUser(user.NickName)
}

func (u *UseCase) CreateUser(user models.User) ([]models.User, error) {
	chUsers, _ := u.repo.CheckUserEmailUniq(user)
	if len(chUsers) > 0 {
		return chUsers, models.Conflict
	}
	usersS := make([]models.User, 0)
	cUser, _ := u.repo.CreateUser(user)
	usersS = append(usersS, cUser)
	return usersS, nil
}
