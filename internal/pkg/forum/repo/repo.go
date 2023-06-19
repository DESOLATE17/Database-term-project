package repo

import (
	"context"
	"fmt"
	"github.com/DESOLATE17/Database-term-project/internal/models"
	"github.com/DESOLATE17/Database-term-project/internal/pkg/forum"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
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

func convertPgErr(err error) error {
	if pqError, ok := err.(*pgconn.PgError); ok {
		switch pqError.Code {
		case "23503": // foreign key
			return models.NotFound
		case "23505": // unique constraint
			return models.Conflict
		default:
			return models.InternalError
		}
	}
	return nil
}

func (r *repoPostgres) GetUser(ctx context.Context, name string) (models.User, error) {
	var userM models.User
	const (
		SelectUserByNickname = `SELECT nickname, fullname, about, email
								FROM users WHERE nickname=$1
								LIMIT 1;`
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
		SelectUserByEmailOrNickname = `SELECT nickname, fullname, about, email
									   FROM users WHERE nickname=$1 OR email=$2
									   LIMIT 2;`
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
		UpdateUserInfo = `UPDATE users
						  SET fullname=coalesce(nullif($1, ''), fullname), about=coalesce(nullif($2, ''), about), email=coalesce(nullif($3, ''), email)
						  WHERE nickname=$4 RETURNING *`
	)
	updatedUser := models.User{}
	row := r.Conn.QueryRow(ctx, UpdateUserInfo, user.FullName, user.About, user.Email, user.NickName)
	err := row.Scan(&updatedUser.NickName, &updatedUser.FullName, &updatedUser.About, &updatedUser.Email)
	if err == pgx.ErrNoRows {
		return updatedUser, models.NotFound
	}
	return updatedUser, convertPgErr(err)
}

func (r *repoPostgres) CreateForum(ctx context.Context, forum models.Forum) error {
	const (
		CreateForum = `INSERT INTO forum(slug, "user", title) VALUES ($1, $2, $3);`
	)
	_, err := r.Conn.Exec(ctx, CreateForum, forum.Slug, forum.User, forum.Title)
	return convertPgErr(err)
}

func (r *repoPostgres) GetForum(ctx context.Context, slug string) (models.Forum, error) {
	const (
		GetForumBySlug = `SELECT title, "user", slug, posts, threads
						  FROM forum WHERE slug=$1
						  LIMIT 1;`
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
		GetThreadBySlug = `SELECT id, title, author, forum, message, votes, slug, created
						   	FROM thread WHERE slug=$1 LIMIT 1;`
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
		GetThreadById = `SELECT id, title, author, forum, message, votes, slug, created
						 FROM thread WHERE id=$1
						 LIMIT 1;`
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
		if post.Parent != 0 {
			old := 0
			query2 := `SELECT thread 
					   FROM post WHERE id = $1;`
			row := r.Conn.QueryRow(ctx, query2, post.Parent)
			err := row.Scan(&old)
			if err != nil || old != thread.ID {
				return []models.Post{}, models.Conflict
			}
		}
	}

	query = strings.TrimSuffix(query, ",")
	query += ` RETURNING id, created, forum, isEdited, thread;`

	rows, err := r.Conn.Query(ctx, query, values...)
	if err != nil {
		return nil, convertPgErr(err)
	}
	defer rows.Close()

	for i := range posts {
		if rows.Next() {
			err = rows.Scan(&posts[i].ID, &posts[i].Created, &posts[i].Forum, &posts[i].IsEdited, &posts[i].Thread)
			if err != nil {
				return nil, convertPgErr(err)
			}
		}
	}
	if rows.Err() != nil {
		return nil, convertPgErr(rows.Err())
	}
	return posts, convertPgErr(err)
}
func (r *repoPostgres) CreateThread(ctx context.Context, thread models.Thread) (models.Thread, error) {
	const (
		InsertThread = `INSERT INTO thread (author, message, title, created, forum, slug, votes)
						VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id;`
	)
	row := r.Conn.QueryRow(ctx, InsertThread, thread.Author, thread.Message, thread.Title,
		thread.Created, thread.Forum, thread.Slug, 0)
	err := row.Scan(&thread.ID)
	return thread, convertPgErr(err)
}

func (r *repoPostgres) GetPostsFlat(ctx context.Context, params models.SortParams, threadID int) ([]models.Post, error) {
	var rows pgx.Rows
	query := `SELECT id, author, created, forum, isedited, message, parent, thread
			  FROM post
		      WHERE thread = $1 `

	if params.Limit == "" && params.Since == "" {
		if params.Desc == "true" {
			query += ` ORDER BY id DESC`
		} else {
			query += ` ORDER BY id ASC`
		}
		rows, _ = r.Conn.Query(ctx, query, threadID)
	} else {
		if params.Limit != "" && params.Since == "" {
			if params.Desc == "true" {
				query += ` ORDER BY id DESC LIMIT $2`
			} else {
				query += `ORDER BY id ASC LIMIT $2`
			}
			rows, _ = r.Conn.Query(ctx, query, threadID, params.Limit)
		}

		if params.Limit != "" && params.Since != "" {
			if params.Desc == "true" {
				query += `AND id < $2 ORDER BY id DESC LIMIT $3`
			} else {
				query += `AND id > $2 ORDER BY id ASC LIMIT $3`
			}
			rows, _ = r.Conn.Query(ctx, query, threadID, params.Since, params.Limit)
		}

		if params.Limit == "" && params.Since != "" {
			if params.Desc == "true" {
				query += `AND id < $2 ORDER BY id DESC`
			} else {
				query += `AND id > $2 ORDER BY id ASC`
			}
			rows, _ = r.Conn.Query(ctx, query, threadID, params.Since)
		}
	}
	posts := make([]models.Post, 0)
	defer rows.Close()
	for rows.Next() {
		onePost := models.Post{}
		err := rows.Scan(&onePost.ID, &onePost.Author, &onePost.Created, &onePost.Forum, &onePost.IsEdited, &onePost.Message, &onePost.Parent, &onePost.Thread)
		if err != nil {
			return posts, models.InternalError
		}
		posts = append(posts, onePost)
	}
	return posts, nil
}

func (r *repoPostgres) GetPostsTree(ctx context.Context, params models.SortParams, threadID int) ([]models.Post, error) {

	var rows pgx.Rows

	query := `SELECT id, author, created, forum, isedited, message, parent, thread
			  FROM post
			  WHERE thread = $1 `

	if params.Limit == "" && params.Since == "" {
		if params.Desc == "true" {
			query += `ORDER BY path, id DESC`
		} else {
			query += `ORDER BY path, id ASC`
		}
		rows, _ = r.Conn.Query(ctx, query, threadID)
	} else {
		if params.Limit != "" && params.Since == "" {
			if params.Desc == "true" {
				query += ` ORDER BY path DESC, id DESC LIMIT $2`
			} else {
				query += ` ORDER BY path, id ASC LIMIT $2`
			}
			rows, _ = r.Conn.Query(ctx, query, threadID, params.Limit)
		}

		if params.Limit != "" && params.Since != "" {
			if params.Desc == "true" {
				query = `SELECT post.id, post.author, post.created, 
				post.forum, post.isedited, post.message, post.parent, post.thread
				FROM post JOIN post parent ON parent.id = $2 WHERE post.path < parent.path AND  post.thread = $1
				ORDER BY post.path DESC, post.id DESC LIMIT $3`
			} else {
				query = `SELECT post.id, post.author, post.created, 
				post.forum, post.isedited, post.message, post.parent, post.thread
				FROM post JOIN post parent ON parent.id = $2 WHERE post.path > parent.path AND  post.thread = $1
				ORDER BY post.path ASC, post.id ASC LIMIT $3`
			}
			rows, _ = r.Conn.Query(ctx, query, threadID, params.Since, params.Limit)
		}

		if params.Limit == "" && params.Since != "" {
			if params.Desc == "true" {
				query = `SELECT post.id, post.author, post.created, 
				post.forum, post.isedited, post.message, post.parent, post.thread
				FROM post JOIN post parent ON parent.id = $2 WHERE post.path < parent.path AND  post.thread = $1
				ORDER BY post.path DESC, post.id DESC`
			} else {
				query = `SELECT post.id, post.author, post.created, 
				post.forum, post.isedited, post.message, post.parent, post.thread
				FROM post JOIN post parent ON parent.id = $2 WHERE post.path > parent.path AND  post.thread = $1
				ORDER BY post.path ASC, post.id ASC`
			}
			rows, _ = r.Conn.Query(ctx, query, threadID, params.Since)
		}
	}

	posts := make([]models.Post, 0)
	defer rows.Close()
	for rows.Next() {
		onePost := models.Post{}
		err := rows.Scan(&onePost.ID, &onePost.Author, &onePost.Created, &onePost.Forum, &onePost.IsEdited, &onePost.Message, &onePost.Parent, &onePost.Thread)
		if err != nil {
			fmt.Println(err)
			return posts, models.InternalError
		}
		posts = append(posts, onePost)
	}

	return posts, nil
}

func (r *repoPostgres) GetPostsParent(ctx context.Context, params models.SortParams, threadID int) ([]models.Post, error) {
	var rows pgx.Rows

	parents := fmt.Sprintf(`SELECT id FROM post WHERE thread = %d AND parent = 0 `, threadID)

	if params.Since != "" {
		if params.Desc == "true" {
			parents += ` AND path[1] < ` + fmt.Sprintf(`(SELECT path[1] FROM post WHERE id = %s) `, params.Since)
		} else {
			parents += ` AND path[1] > ` + fmt.Sprintf(`(SELECT path[1] FROM post WHERE id = %s) `, params.Since)
		}
	}

	if params.Desc == "true" {
		parents += ` ORDER BY id DESC `
	} else {
		parents += ` ORDER BY id ASC `
	}

	if params.Limit != "" {
		parents += " LIMIT " + params.Limit
	}

	query := fmt.Sprintf(
		`SELECT id, author, created, forum, isedited, message, parent, thread FROM post WHERE path[1] = ANY (%s) `, parents)

	if params.Desc == "true" {
		query += ` ORDER BY path[1] DESC, path,  id `
	} else {
		query += ` ORDER BY path[1] ASC, path,  id `
	}

	rows, _ = r.Conn.Query(ctx, query)
	posts := make([]models.Post, 0)
	defer rows.Close()
	for rows.Next() {
		onePost := models.Post{}
		err := rows.Scan(&onePost.ID, &onePost.Author, &onePost.Created, &onePost.Forum, &onePost.IsEdited, &onePost.Message, &onePost.Parent, &onePost.Thread)
		if err != nil {
			fmt.Println(err)
			return posts, models.InternalError
		}
		posts = append(posts, onePost)
	}

	return posts, nil
}

func (r *repoPostgres) GetForumThreads(ctx context.Context, forum models.Forum, params models.SortParams) ([]models.Thread, error) {
	var rows pgx.Rows
	var err error
	threads := make([]models.Thread, 0)
	const (
		GetThreadsSinceDescNotNil = `SELECT id, title, author, forum, message, votes, slug, created
                                     	FROM thread WHERE forum=$1 AND created <= $2 
                                        ORDER BY created DESC LIMIT $3;`
		GetThreadsSinceDescNil = `SELECT id, title, author, forum, message, votes, slug, created
                                     FROM thread WHERE forum=$1 AND created >= $2
                                     ORDER BY created ASC  LIMIT $3;`
		GetThreadsDescNotNil = `SELECT id, title, author, forum, message, votes, slug, created
                                	FROM thread WHERE forum=$1
                                	ORDER BY created DESC LIMIT $2;`
		GetThreadsDescNil = `SELECT id, title, author, forum, message, votes, slug, created
									FROM thread WHERE forum=$1
			             			ORDER BY created ASC  LIMIT $2;`
	)

	if params.Since != "" {
		if params.Desc == "true" {
			rows, err = r.Conn.Query(ctx, GetThreadsSinceDescNotNil, forum.Slug, params.Since, params.Limit)
		} else {
			rows, err = r.Conn.Query(ctx, GetThreadsSinceDescNil, forum.Slug, params.Since, params.Limit)
		}
	} else {
		if params.Desc == "true" {
			rows, err = r.Conn.Query(ctx, GetThreadsDescNotNil, forum.Slug, params.Limit)
		} else {
			rows, err = r.Conn.Query(ctx, GetThreadsDescNil, forum.Slug, params.Limit)
		}
	}

	if err != nil {
		return threads, models.NotFound
	}
	defer rows.Close()
	for rows.Next() {
		threadS := models.Thread{}
		err = rows.Scan(&threadS.ID, &threadS.Title, &threadS.Author, &threadS.Forum, &threadS.Message,
			&threadS.Votes, &threadS.Slug, &threadS.Created)
		if err != nil {
			continue
		}
		threads = append(threads, threadS)
	}
	return threads, nil
}

func (r *repoPostgres) ForumCheck(ctx context.Context, slug string) (string, error) {
	const (
		SelectSlugFromForum = `SELECT slug
							   FROM forum
							   WHERE slug = $1;`
	)
	row := r.Conn.QueryRow(ctx, SelectSlugFromForum, slug)
	err := row.Scan(&slug)
	if err != nil {
		return slug, models.NotFound
	}
	return slug, nil
}

func (r *repoPostgres) Vote(ctx context.Context, vote models.Vote) error {
	const (
		CreateVote = `INSERT INTO vote(author, voice, thread)
					  VALUES ($1, $2, $3);`
	)

	_, err := r.Conn.Exec(ctx, CreateVote, vote.Nickname, vote.Voice, vote.Thread)
	return convertPgErr(err)
}

func (r *repoPostgres) UpdateVote(ctx context.Context, vote models.Vote) error {
	const (
		UpdateVote = `UPDATE vote SET voice=$1 WHERE author=$2 AND thread=$3;`
	)

	_, err := r.Conn.Exec(ctx, UpdateVote, vote.Voice, vote.Nickname, vote.Thread)
	if err != nil {
		return err
	}
	return nil
}

func (r *repoPostgres) UpdateThreadInfo(ctx context.Context, upThread models.Thread) (models.Thread, error) {
	threadS := models.Thread{}
	var row pgx.Row
	const (
		UpdateThread = `UPDATE thread SET title=coalesce(nullif($1, ''), title), message=coalesce(nullif($2, ''), message)
                        WHERE %s RETURNING *;`
	)

	if upThread.Slug == "" {
		rowQuery := fmt.Sprintf(UpdateThread, `id=$3`)
		row = r.Conn.QueryRow(ctx, rowQuery, upThread.Title, upThread.Message, upThread.ID)
	} else {
		rowQuery := fmt.Sprintf(UpdateThread, `slug=$3`)
		row = r.Conn.QueryRow(ctx, rowQuery, upThread.Title, upThread.Message, upThread.Slug)
	}
	err := row.Scan(&threadS.ID, &threadS.Title, &threadS.Author,
		&threadS.Forum, &threadS.Message, &threadS.Votes, &threadS.Slug, &threadS.Created)
	if err != nil {
		return models.Thread{}, models.NotFound
	}
	return threadS, nil
}

func (r *repoPostgres) GetUsersOfForum(ctx context.Context, forum models.Forum, params models.SortParams) ([]models.User, error) {
	var query string
	const (
		GetUsersOfForumDescNotNilSince = `SELECT nickname, fullname, about, email
										  FROM users_forum 
										  WHERE slug=$1 AND nickname < '%s' 
										  ORDER BY nickname DESC 
										  LIMIT nullif($2, 0)`
		GetUsersOfForumDescSinceNil = `SELECT nickname, fullname, about, email
										  FROM users_forum
										  WHERE slug=$1 
										  ORDER BY nickname DESC
										  LIMIT nullif($2, 0)`
		GetUsersOfForumAsc = `SELECT nickname, fullname, about, email
										  FROM users_forum 
										  WHERE slug=$1 AND nickname > '%s' 
										  ORDER BY nickname 
										  LIMIT nullif($2, 0)`
	)

	if params.Desc == "true" {
		if params.Since != "" {
			query = fmt.Sprintf(GetUsersOfForumDescNotNilSince, params.Since)
		} else {
			query = GetUsersOfForumDescSinceNil
		}
	} else {
		query = fmt.Sprintf(GetUsersOfForumAsc, params.Since)
	}
	users := make([]models.User, 0)
	rows, err := r.Conn.Query(ctx, query, forum.Slug, params.Limit)

	if err != nil {
		return users, models.NotFound
	}

	defer rows.Close()

	for rows.Next() {
		user := models.User{}
		err = rows.Scan(&user.NickName, &user.FullName, &user.About, &user.Email)
		if err != nil {
			return users, models.InternalError
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *repoPostgres) GetFullPostInfo(ctx context.Context, posts models.PostFull, related []string) (models.PostFull, error) {
	post := models.Post{}
	postFull := models.PostFull{}

	const (
		SelectPostById = `SELECT author, message, created, forum, isedited, parent, thread
						  FROM post WHERE id = $1;`
	)

	post.ID = posts.Post.ID

	row := r.Conn.QueryRow(ctx, SelectPostById, posts.Post.ID)
	err := row.Scan(&post.Author, &post.Message, &post.Created, &post.Forum, &post.IsEdited, &post.Parent, &post.Thread)
	if err != nil {
		return postFull, models.NotFound
	}

	postFull.Post = post

	for i := 0; i < len(related); i++ {
		if "user" == related[i] {
			user, _ := r.GetUser(ctx, post.Author)
			postFull.Author = &user
		}
		if "forum" == related[i] {
			forumS, _ := r.GetForum(ctx, post.Forum)
			postFull.Forum = &forumS
		}
		if "thread" == related[i] {
			thread, _ := r.GetThreadByID(ctx, post.Thread)
			postFull.Thread = &thread

		}
	}
	return postFull, nil
}

func (r *repoPostgres) UpdatePostInfo(ctx context.Context, postUpdate models.PostUpdate) (models.Post, error) {
	const (
		UpdatePostMessage = `UPDATE post SET message=coalesce(nullif($1, ''), message), isedited = CASE WHEN $1 = '' OR message = $1 THEN isedited ELSE TRUE END WHERE id=$2 RETURNING *`
	)
	postOne := models.Post{}
	row := r.Conn.QueryRow(ctx, UpdatePostMessage, postUpdate.Message, postUpdate.ID)
	err := row.Scan(&postOne.ID, &postOne.Author, &postOne.Created, &postOne.Forum,
		&postOne.IsEdited, &postOne.Message, &postOne.Parent, &postOne.Thread, &postOne.Path)
	if err != nil {
		fmt.Println(err)
		return postOne, models.NotFound
	}
	return postOne, nil
}

// TODO maybe should do 1 query
func (r *repoPostgres) GetStatus(ctx context.Context) models.Status {
	const (
		GetStatus = `SELECT Threads, Users, Forums, Posts FROM status;`
	)
	status := models.Status{}
	row := r.Conn.QueryRow(ctx, GetStatus)
	err := row.Scan(&status.Threads, &status.Users, &status.Forums, &status.Posts)
	fmt.Println(err)
	return status
}

func (r *repoPostgres) GetClear(ctx context.Context) {
	const (
		ClearAll = `TRUNCATE TABLE users, forum, thread, post, vote, users_forum, status CASCADE;`
	)
	_, _ = r.Conn.Exec(ctx, ClearAll)
}
