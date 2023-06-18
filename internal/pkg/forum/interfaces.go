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
	GetPostOfThread(ctx context.Context, params models.SortParams, threadID int) ([]models.Post, error)
	GetForumThreads(ctx context.Context, forum models.Forum, params models.SortParams) ([]models.Thread, error)
	Vote(ctx context.Context, vote models.Vote) error
	UpdateThreadInfo(ctx context.Context, slugOrId string, updateThread models.Thread) (models.Thread, error)
	GetUsersOfForum(ctx context.Context, forum models.Forum, params models.SortParams) ([]models.User, error)
	GetFullPostInfo(ctx context.Context, posts models.PostFull, related []string) (models.PostFull, error)
	UpdatePostInfo(ctx context.Context, postUpdate models.PostUpdate) (models.Post, error)
	GetStatus(ctx context.Context) models.Status
	GetClear(ctx context.Context)
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
	GetPostsFlat(ctx context.Context, params models.SortParams, threadID int) ([]models.Post, error)
	GetPostsTree(ctx context.Context, params models.SortParams, threadID int) ([]models.Post, error)
	GetPostsParent(ctx context.Context, params models.SortParams, threadID int) ([]models.Post, error)
	GetForumThreads(ctx context.Context, forum models.Forum, params models.SortParams) ([]models.Thread, error)
	ForumCheck(ctx context.Context, slug string) (string, error)
	Vote(ctx context.Context, vote models.Vote) error
	UpdateVote(ctx context.Context, vote models.Vote) error
	UpdateThreadInfo(ctx context.Context, upThread models.Thread) (models.Thread, error)
	GetUsersOfForum(ctx context.Context, forum models.Forum, params models.SortParams) ([]models.User, error)
	GetFullPostInfo(ctx context.Context, posts models.PostFull, related []string) (models.PostFull, error)
	UpdatePostInfo(ctx context.Context, postUpdate models.PostUpdate) (models.Post, error)
	GetStatus(ctx context.Context) models.Status
	GetClear(ctx context.Context)
}
