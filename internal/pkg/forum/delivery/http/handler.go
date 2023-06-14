package handler

import (
	"encoding/json"
	"github.com/DESOLATE17/Database-term-project/internal/models"
	"github.com/DESOLATE17/Database-term-project/internal/pkg/forum"
	"github.com/DESOLATE17/Database-term-project/internal/utils"
	"github.com/gorilla/mux"
	"github.com/mailru/easyjson"
	"net/http"
)

type Handler struct {
	uc forum.UseCase
}

func NewForumHandler(ForumUseCase forum.UseCase) *Handler {
	return &Handler{uc: ForumUseCase}
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nickname, found := vars["nickname"]
	if !found {
		utils.Response(w, http.StatusNotFound, nil)
		return
	}

	userS := models.User{}
	userS.NickName = nickname

	finalUser, err := h.uc.GetUser(r.Context(), userS)
	if err != nil {
		utils.Response(w, http.StatusNotFound, nickname)
		return
	}
	utils.Response(w, http.StatusOK, finalUser)
	return
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nickname, found := vars["nickname"]
	if !found {
		utils.Response(w, http.StatusNotFound, nil)
		return
	}

	user := models.User{}
	err := easyjson.UnmarshalFromReader(r.Body, &user)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, nil)
		return
	}
	user.NickName = nickname

	finalUser, err := h.uc.CreateUser(r.Context(), user)
	if err == nil {
		newU := finalUser[0]
		utils.Response(w, http.StatusCreated, newU)
		return
	}
	utils.Response(w, http.StatusConflict, finalUser)
}

func (h *Handler) ChangeUserInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nickname, found := vars["nickname"]
	if !found {
		utils.Response(w, http.StatusNotFound, nil)
		return
	}

	user := models.User{}
	err := easyjson.UnmarshalFromReader(r.Body, &user)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, nil)
		return
	}
	user.NickName = nickname

	updatedUser, err := h.uc.UpdateUserInfo(r.Context(), user)
	if err == nil {
		utils.Response(w, http.StatusOK, updatedUser)
		return
	}
	if err == models.NotFound {
		utils.Response(w, http.StatusNotFound, nickname)
		return
	}
	utils.Response(w, http.StatusConflict, models.ErrorResponse{Message: nickname})
}

func (h *Handler) CreateForum(w http.ResponseWriter, r *http.Request) {
	forum := models.Forum{}
	err := easyjson.UnmarshalFromReader(r.Body, &forum)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, nil)
		return
	}

	finalForum, err := h.uc.CreateForum(r.Context(), forum)
	if err == models.Conflict {
		utils.Response(w, http.StatusConflict, finalForum)
		return
	}
	if err == models.NotFound {
		utils.Response(w, http.StatusNotFound, forum.User)
		return
	}
	utils.Response(w, http.StatusCreated, finalForum)
}

func (h *Handler) ForumInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug, found := vars["slug"]
	if !found {
		utils.Response(w, http.StatusNotFound, nil)
		return
	}

	forum, err := h.uc.ForumInfo(r.Context(), slug)
	if err == models.NotFound {
		utils.Response(w, http.StatusNotFound, slug)
		return
	}
	utils.Response(w, http.StatusOK, forum)
}

func (h *Handler) CreatePosts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slugOrId, found := vars["slug_or_id"]
	if !found {
		utils.Response(w, http.StatusNotFound, nil)
		return
	}

	thread, err := h.uc.CheckThreadIdOrSlug(r.Context(), slugOrId)
	if err != nil {
		utils.Response(w, http.StatusNotFound, slugOrId)
		return
	}
	var posts []models.Post
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&posts)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, nil)
		return
	}

	if len(posts) == 0 {
		utils.Response(w, http.StatusCreated, []models.Post{})
		return
	}

	posts, err = h.uc.CreatePosts(r.Context(), posts, thread)
	if err == models.NotFound {
		utils.Response(w, http.StatusNotFound, slugOrId)
		return
	}
	if err == models.Conflict {
		utils.Response(w, http.StatusConflict, nil)
		return
	}

	utils.Response(w, http.StatusCreated, posts)
}

func (h *Handler) CreateForumThread(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug, found := vars["slug"]
	if !found {
		utils.Response(w, http.StatusNotFound, nil)
		return
	}

	thread := models.Thread{}
	err := easyjson.UnmarshalFromReader(r.Body, &thread)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, nil)
		return
	}
	thread.Forum = slug

	thread, err = h.uc.CreateForumThread(r.Context(), thread)
	if err == models.Conflict {
		utils.Response(w, http.StatusConflict, thread)
		return
	}
	if err == models.NotFound {
		utils.Response(w, http.StatusNotFound, models.ErrorResponse{Message: slug})
		return
	}
	utils.Response(w, http.StatusCreated, thread)
}
