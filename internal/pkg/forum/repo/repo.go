package repo

import (
	"context"
	"github.com/DESOLATE17/Database-term-project/internal/models"
	"github.com/DESOLATE17/Database-term-project/internal/pkg/forum"
	"github.com/jackc/pgx/v4/pgxpool"
)

type repoPostgres struct {
	Conn *pgxpool.Pool
}

func NewRepoPostgres(Conn *pgxpool.Pool) forum.Repository {
	return &repoPostgres{Conn: Conn}
}

func (r *repoPostgres) GetUser(name string) (models.User, error) {
	var userM models.User
	const SelectUserByNickname = "select nickname, fullname, about, email from users where nickname=$1 limit 1;"
	row := r.Conn.QueryRow(context.Background(), SelectUserByNickname, name)
	err := row.Scan(&userM.NickName, &userM.FullName, &userM.About, &userM.Email)
	if err != nil {
		return models.User{}, models.NotFound
	}
	return userM, nil
}

func (r *repoPostgres) CheckUserEmailUniq(usersS models.User) ([]models.User, error) {
	const SelectUserByEmailOrNickname = "select nickname, fullname, about, email from users where nickname=$1 or email=$2 limit 2;"
	rows, err := r.Conn.Query(context.Background(), SelectUserByEmailOrNickname, usersS.NickName, usersS.Email)
	defer rows.Close()
	if err != nil {
		return []models.User{}, models.InternalError
	}
	users := make([]models.User, 0)
	for rows.Next() {
		userOne := models.User{}
		err := rows.Scan(&userOne.NickName, &userOne.FullName, &userOne.About, &userOne.Email)
		if err != nil {
			return []models.User{}, models.InternalError
		}
		users = append(users, userOne)
	}
	return users, nil
}

func (r *repoPostgres) CreateUser(user models.User) (models.User, error) {
	_, err := r.Conn.Exec(context.Background(), `Insert INTO users(Nickname, FullName, About, Email) VALUES ($1, $2, $3, $4);`,
		user.NickName, user.FullName, user.About, user.Email)
	if err != nil {
		return models.User{}, models.InternalError
	}
	return user, nil
}
