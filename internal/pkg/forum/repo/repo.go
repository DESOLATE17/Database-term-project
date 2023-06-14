package repo

import (
	"context"
	"fmt"
	"github.com/DESOLATE17/Database-term-project/internal/models"
	"github.com/DESOLATE17/Database-term-project/internal/pkg/forum"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
	"strings"
	"time"
)

type repoPostgres struct {
	Conn *pgxpool.Pool
}

func NewRepoPostgres(Conn *pgxpool.Pool) forum.Repository {
	return &repoPostgres{Conn: Conn}
}

func (r *repoPostgres) GetUser(ctx context.Context, name string) (models.User, error) {
	var userM models.User
	const (
		SelectUserByNickname = "SELECT nickname, fullname, about, email FROM users WHERE nickname=$1 LIMIT 1;"
	)
	row := r.Conn.QueryRow(ctx, SelectUserByNickname, name)
	err := row.Scan(&userM.NickName, &userM.FullName, &userM.About, &userM.Email)
	if err != nil {
		return models.User{}, models.NotFound
	}
	return userM, nil
}

func (r *repoPostgres) CheckUserEmailAndNicknameUniq(ctx context.Context, user models.User) ([]models.User, error) {
	const (
		SelectUserByEmailOrNickname = `SELECT nickname, fullname, about, email FROM users WHERE nickname=$1 OR email=$2 LIMIT 2;`
	)
	rows, err := r.Conn.Query(ctx, SelectUserByEmailOrNickname, user.NickName, user.Email)
	defer rows.Close()
	if err != nil {
		return []models.User{}, models.InternalError
	}
	users := make([]models.User, 0, 2)
	for rows.Next() {
		user := models.User{}
		err = rows.Scan(&user.NickName, &user.FullName, &user.About, &user.Email)
		if err != nil {
			return []models.User{}, models.InternalError
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *repoPostgres) CreateUser(ctx context.Context, user models.User) error {
	const CreateUser = `INSERT INTO users(Nickname, FullName, About, Email) VALUES ($1, $2, $3, $4);`
	_, err := r.Conn.Exec(ctx, CreateUser, user.NickName, user.FullName, user.About, user.Email)
	if err != nil {
		return models.InternalError
	}
	return nil
}
func (r *repoPostgres) UpdateUserInfo(ctx context.Context, user models.User) (models.User, error) {
	const (
		UpdateUserInfo = `UPDATE users SET fullname=coalesce(nullif($1, ''), fullname), about=coalesce(nullif($2, ''), about), email=coalesce(nullif($3, ''), email) WHERE nickname=$4 RETURNING *`
	)
	updatedUser := models.User{}
	row := r.Conn.QueryRow(ctx, UpdateUserInfo, user.FullName, user.About, user.Email, user.NickName)
	err := row.Scan(&updatedUser.NickName, &updatedUser.FullName, &updatedUser.About, &updatedUser.Email)
	if err != nil {
		return updatedUser, err
	}
	return updatedUser, nil
}

func (r *repoPostgres) CreateForum(ctx context.Context, forum models.Forum) error {
	const (
		CreateForum = `INSERT INTO forum(slug, "user", title) VALUES ($1, $2, $3);`
	)
	_, err := r.Conn.Exec(ctx, CreateForum, forum.Slug, forum.User, forum.Title)
	if err != nil {
		return err
	}
	return nil
}

func (r *repoPostgres) GetForum(ctx context.Context, slug string) (models.Forum, error) {
	const (
		GetForumBySlug = `SELECT title, "user", slug, posts, threads FROM forum WHERE slug=$1 LIMIT 1;`
	)
	forum := models.Forum{}
	row := r.Conn.QueryRow(ctx, GetForumBySlug, slug)
	err := row.Scan(&forum.Title, &forum.User, &forum.Slug, &forum.Posts, &forum.Threads)
	if err != nil {
		return forum, models.NotFound
	}
	return forum, nil
}

func (r *repoPostgres) GetThreadBySlug(ctx context.Context, slug string) (models.Thread, error) {
	thread := models.Thread{}
	const (
		GetThreadBySlug = `SELECT id, title, author, forum, message, votes, slug, created FROM thread WHERE slug=$1 LIMIT 1;`
	)
	row := r.Conn.QueryRow(ctx, GetThreadBySlug, slug)
	err := row.Scan(&thread.ID, &thread.Title, &thread.Author, &thread.Forum,
		&thread.Message, &thread.Votes, &thread.Slug, &thread.Created)
	if err != nil {
		return models.Thread{}, models.NotFound
	}
	return thread, nil
}

func (r *repoPostgres) GetThreadByID(ctx context.Context, id int) (models.Thread, error) {
	thread := models.Thread{}
	const (
		GetThreadById = `SELECT id, title, author, forum, message, votes, slug, created FROM thread WHERE id=$1 LIMIT 1;`
	)

	row := r.Conn.QueryRow(ctx, GetThreadById, id)

	err := row.Scan(&thread.ID, &thread.Title, &thread.Author, &thread.Forum,
		&thread.Message, &thread.Votes, &thread.Slug, &thread.Created)
	if err != nil {
		return models.Thread{}, models.NotFound
	}
	return thread, nil
}

func (r *repoPostgres) CreatePosts(ctx context.Context, posts []models.Post, thread models.Thread) ([]models.Post, error) {
	query := "INSERT INTO post(author, created, forum, message, parent, thread) VALUES "
	values := make([]interface{}, 0)
	created := time.Now()

	for i, post := range posts {
		query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d),", i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6)
		values = append(values, post.Author, created, thread.Forum, post.Message, post.Parent, thread.ID)
	}

	query = strings.TrimSuffix(query, ",")
	query += ` RETURNING id, created, forum, isEdited, thread;`

	rows, err := r.Conn.Query(ctx, query, values...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for i := range posts {
		if rows.Next() {
			err = rows.Scan(&posts[i].ID, &posts[i].Created, &posts[i].Forum, &posts[i].IsEdited, &posts[i].Thread)
			if err != nil {
				return nil, err
			}
		}
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return posts, nil
}
func (r *repoPostgres) CreateThread(ctx context.Context, thread models.Thread) (models.Thread, error) {
	const (
		InsertThread = `INSERT INTO thread (author, message, title, created, forum, slug, votes) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id;`
	)
	row := r.Conn.QueryRow(ctx, InsertThread, thread.Author, thread.Message, thread.Title,
		thread.Created, thread.Forum, thread.Slug, 0)
	err := row.Scan(&thread.ID)

	if err != nil {
		if pqError, ok := err.(*pgconn.PgError); ok {
			switch pqError.Code {
			case "23503": // foreign key
				return models.Thread{}, models.NotFound
			case "23505": // unique constraint
				return thread, models.Conflict
			default:
				return models.Thread{}, models.InternalError
			}
		}
	}
	return thread, nil
}
