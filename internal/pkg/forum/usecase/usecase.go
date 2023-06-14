package usecase

import (
	"context"
	"github.com/DESOLATE17/Database-term-project/internal/models"
	"github.com/DESOLATE17/Database-term-project/internal/pkg/forum"
	"github.com/jackc/pgconn"
	"strconv"
)

type UseCase struct {
	repo forum.Repository
}

func NewRepoUsecase(repo forum.Repository) forum.UseCase {
	return &UseCase{repo: repo}
}

func (u *UseCase) GetUser(ctx context.Context, user models.User) (models.User, error) {
	return u.repo.GetUser(ctx, user.NickName)
}

func (u *UseCase) CreateUser(ctx context.Context, user models.User) ([]models.User, error) {
	usersWithSameInfo, _ := u.repo.CheckUserEmailAndNicknameUniq(ctx, user)
	if len(usersWithSameInfo) > 0 {
		return usersWithSameInfo, models.Conflict
	}
	err := u.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return []models.User{user}, nil
}

func (u *UseCase) UpdateUserInfo(ctx context.Context, user models.User) (models.User, error) {
	updatedUser, err := u.repo.UpdateUserInfo(ctx, user)
	if err != nil {
		if pgError, ok := err.(*pgconn.PgError); ok && pgError.Code == "23505" { // unique_violation code
			return updatedUser, models.Conflict
		}
		return updatedUser, models.NotFound
	}
	return updatedUser, nil
}

func (u *UseCase) CreateForum(ctx context.Context, forum models.Forum) (models.Forum, error) {
	user, err := u.repo.GetUser(ctx, forum.User)
	if err != nil {
		return models.Forum{}, err
	}
	forum.User = user.NickName

	err = u.repo.CreateForum(ctx, forum)
	if err != nil {
		//if pgError, ok := errI.(*pgconn.PgError); ok && pgError.Code == "23503" {
		//	return models.Forum{}, models.NotFound
		//}
		if pgError, ok := err.(*pgconn.PgError); ok && pgError.Code == "23505" {
			forum, _ = u.repo.GetForum(ctx, forum.Slug)
			return forum, models.Conflict
		}
		return models.Forum{}, models.InternalError
	}
	return forum, nil
}

func (u *UseCase) ForumInfo(ctx context.Context, slug string) (models.Forum, error) {
	return u.repo.GetForum(ctx, slug)
}

func (u *UseCase) CheckThreadIdOrSlug(ctx context.Context, slugOrId string) (models.Thread, error) {
	threadID, err := strconv.Atoi(slugOrId)
	if err != nil {
		return u.repo.GetThreadBySlug(ctx, slugOrId)
	}
	return u.repo.GetThreadByID(ctx, threadID)
}

func (u *UseCase) CreatePosts(ctx context.Context, posts []models.Post, thread models.Thread) ([]models.Post, error) {
	posts, err := u.repo.CreatePosts(ctx, posts, thread)
	if err != nil {
		if pgError, ok := err.(*pgconn.PgError); ok && pgError.Code == "23503" { // foreign key
			return nil, models.NotFound
		}
		return nil, models.Conflict
	}
	return posts, nil
}

func (u *UseCase) CreateForumThread(ctx context.Context, thread models.Thread) (models.Thread, error) {
	//user, err := u.repo.GetUser(ctx, thread.Author)
	//if err != nil {
	//	return models.Thread{}, models.NotFound
	//}
	//
	//forum, err := u.repo.ForumCheck(models.Forum{Slug: thread.Forum})
	//if err == models.NotFound {
	//	return models.Thread{}, models.NotFound
	//}
	//
	//thread.Author = user.NickName
	//thread.Forum = forum.Slug

	//if thread.Slug != "" {
	//	thread, status := u.repo.CheckSlug(thread)
	//	if status == nil {
	//		th, _ := u.repo.GetThreadBySlug(ctx, thread.Slug)
	//		return th, models.Conflict
	//	}
	//}

	return u.repo.CreateThread(ctx, thread)
}
