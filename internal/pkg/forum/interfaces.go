package forum

import (
	"context"
	"github.com/DESOLATE17/Database-term-project/internal/models"
)

type UseCase interface {
	GetUser(ctx context.Context, user models.User) (models.User, error)
	CreateUser(ctx context.Context, user models.User) ([]models.User, error)
	UpdateUserInfo(ctx context.Context, user models.User) (models.User, error)
	CreateForum(ctx context.Context, forum models.Forum) (models.Forum, error)
	ForumInfo(ctx context.Context, slug string) (models.Forum, error)
	CheckThreadIdOrSlug(ctx context.Context, slugOrId string) (models.Thread, error)
	CreatePosts(ctx context.Context, posts []models.Post, thread models.Thread) ([]models.Post, error)
	CreateForumThread(ctx context.Context, thread models.Thread) (models.Thread, error)
}

type Repository interface {
	GetUser(ctx context.Context, name string) (models.User, error)
	CheckUserEmailAndNicknameUniq(ctx context.Context, user models.User) ([]models.User, error)
	CreateUser(ctx context.Context, user models.User) error
	UpdateUserInfo(ctx context.Context, user models.User) (models.User, error)
	CreateForum(ctx context.Context, forum models.Forum) error
	GetForum(ctx context.Context, slug string) (models.Forum, error)
	GetThreadBySlug(ctx context.Context, slug string) (models.Thread, error)
	GetThreadByID(ctx context.Context, id int) (models.Thread, error)
	CreatePosts(ctx context.Context, posts []models.Post, thread models.Thread) ([]models.Post, error)
	CreateThread(ctx context.Context, thread models.Thread) (models.Thread, error)
}
