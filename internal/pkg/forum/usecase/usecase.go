package usecase

import (
	"context"
	"github.com/DESOLATE17/Database-term-project/internal/models"
	"github.com/DESOLATE17/Database-term-project/internal/pkg/forum"
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
	return u.repo.UpdateUserInfo(ctx, user)
}

func (u *UseCase) CreateForum(ctx context.Context, forum models.Forum) (models.Forum, error) {
	user, err := u.repo.GetUser(ctx, forum.User)
	if err != nil {
		return models.Forum{}, err
	}
	forum.User = user.NickName

	err = u.repo.CreateForum(ctx, forum)
	if err == models.Conflict {
		forum, _ = u.repo.GetForum(ctx, forum.Slug)
		return forum, models.Conflict
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
	return u.repo.CreatePosts(ctx, posts, thread)
}

func (u *UseCase) CreateForumThread(ctx context.Context, thread models.Thread) (models.Thread, error) {
	if thread.Slug != "" {
		th, err := u.repo.GetThreadBySlug(ctx, thread.Slug)
		if err == nil {
			return th, models.Conflict
		}
	}
	f, status := u.repo.ForumCheck(ctx, thread.Forum)
	if status == models.NotFound {
		return models.Thread{}, models.NotFound
	}
	thread.Forum = f

	return u.repo.CreateThread(ctx, thread)
}

func (u *UseCase) GetPostOfThread(ctx context.Context, params models.SortParams, threadID int) ([]models.Post, error) {
	switch params.Sort {
	case "flat":
		return u.repo.GetPostsFlat(ctx, params, threadID)
	case "tree":
		return u.repo.GetPostsTree(ctx, params, threadID)
	case "parent_tree":
		return u.repo.GetPostsParent(ctx, params, threadID)
	default:
		return u.repo.GetPostsFlat(ctx, params, threadID)
	}
}

func (u *UseCase) GetForumThreads(ctx context.Context, forum models.Forum, params models.SortParams) ([]models.Thread, error) {
	_, err := u.repo.GetForum(ctx, forum.Slug)
	if err != nil {
		return nil, err
	}
	return u.repo.GetForumThreads(ctx, forum, params)
}

func (u *UseCase) Vote(ctx context.Context, vote models.Vote) error {
	err := u.repo.Vote(ctx, vote)
	if err == models.Conflict {
		return u.repo.UpdateVote(ctx, vote)
	}
	return err
}

func (u *UseCase) UpdateThreadInfo(ctx context.Context, slugOrId string, updateThread models.Thread) (models.Thread, error) {
	threadID, err := strconv.Atoi(slugOrId)
	if err != nil {
		updateThread.Slug = slugOrId
	} else {
		updateThread.ID = threadID
	}
	return u.repo.UpdateThreadInfo(ctx, updateThread)
}

func (u *UseCase) GetUsersOfForum(ctx context.Context, forum models.Forum, params models.SortParams) ([]models.User, error) {
	_, err := u.repo.GetForum(ctx, forum.Slug)
	if err != nil {
		return nil, err
	}

	return u.repo.GetUsersOfForum(ctx, forum, params)
}

func (u *UseCase) GetFullPostInfo(ctx context.Context, posts models.PostFull, related []string) (models.PostFull, error) {
	return u.repo.GetFullPostInfo(ctx, posts, related)
}

func (u *UseCase) UpdatePostInfo(ctx context.Context, postUpdate models.PostUpdate) (models.Post, error) {
	return u.repo.UpdatePostInfo(ctx, postUpdate)
}

func (u *UseCase) GetClear(ctx context.Context) {
	u.repo.GetClear(ctx)
}

func (u *UseCase) GetStatus(ctx context.Context) models.Status {
	return u.repo.GetStatus(ctx)
}
