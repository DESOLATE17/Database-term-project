package main

import (
	"context"
	delivery "github.com/DESOLATE17/Database-term-project/internal/pkg/forum/delivery/http"
	"github.com/DESOLATE17/Database-term-project/internal/pkg/forum/repo"
	"github.com/DESOLATE17/Database-term-project/internal/pkg/forum/usecase"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"net/http"
)

// sudo docker rm -f my_container
// sudo docker build -t docker .
// sudo docker run -p 5000:5000 --name my_container -t docker

func main() {
	muxRoute := mux.NewRouter()
	//conn := "postgres://postgres:password@127.0.0.1:5432/bd?sslmode=disable&pool_max_conns=1000"
	conn := "postgres://docker:docker@127.0.0.1:5432/docker?sslmode=disable&pool_max_conns=1000"
	pool, err := pgxpool.Connect(context.Background(), conn)
	if err != nil {
		log.Fatal("No connection to postgres", err)
	}

	fRepo := repo.NewRepoPostgres(pool)
	fUsecase := usecase.NewRepoUsecase(fRepo)
	fHandler := delivery.NewForumHandler(fUsecase)

	forum := muxRoute.PathPrefix("/api").Subrouter()
	{
		forum.HandleFunc("/user/{nickname}/create", fHandler.CreateUser).Methods(http.MethodPost)
		forum.HandleFunc("/user/{nickname}/profile", fHandler.GetUser).Methods(http.MethodGet)
		//forum.HandleFunc("/user/{nickname}/profile", fHandler.ChangeInfoUser).Methods(http.MethodPost)

		//forum.HandleFunc("/forum/create", fHandler.CreateForum).Methods(http.MethodPost)
		//forum.HandleFunc("/forum/{slug}/details", fHandler.ForumInfo).Methods(http.MethodGet)
		//forum.HandleFunc("/forum/{slug}/create", fHandler.CreateThreadsForum).Methods(http.MethodPost)
		//forum.HandleFunc("/forum/{slug}/users", fHandler.GetUsersForum).Methods(http.MethodGet)
		//forum.HandleFunc("/forum/{slug}/threads", fHandler.GetThreadsForum).Methods(http.MethodGet)
		//
		//forum.HandleFunc("/post/{id}/details", fHandler.GetPostInfo).Methods(http.MethodGet)
		//forum.HandleFunc("/post/{id}/details", fHandler.UpdatePostInfo).Methods(http.MethodPost)
		//
		//forum.HandleFunc("/service/clear", fHandler.GetClear).Methods(http.MethodPost)
		//forum.HandleFunc("/service/status", fHandler.GetStatus).Methods(http.MethodGet)
		//
		//forum.HandleFunc("/thread/{slug_or_id}/create", fHandler.CreatePosts).Methods(http.MethodPost)
		//forum.HandleFunc("/thread/{slug_or_id}/details", fHandler.GetThreadInfo).Methods(http.MethodGet)
		//forum.HandleFunc("/thread/{slug_or_id}/details", fHandler.UpdateThreadInfo).Methods(http.MethodPost)
		//forum.HandleFunc("/thread/{slug_or_id}/posts", fHandler.GetPostOfThread).Methods(http.MethodGet)
		//forum.HandleFunc("/thread/{slug_or_id}/vote", fHandler.Voted).Methods(http.MethodPost)
	}

	http.Handle("/", muxRoute)
	log.Print(http.ListenAndServe(":5000", muxRoute))
}
