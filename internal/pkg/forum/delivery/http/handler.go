package handler

import (
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

	finalUser, _ := h.uc.GetUser(userS)
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

	userS := models.User{}
	err := easyjson.UnmarshalFromReader(r.Body, &userS)
	if err != nil {
		utils.Response(w, http.StatusInternalServerError, nil)
		return
	}
	userS.NickName = nickname

	finalUser, err := h.uc.CreateUser(userS)
	if err == nil {
		newU := finalUser[0]
		utils.Response(w, http.StatusCreated, newU)
		return
	}
	utils.Response(w, http.StatusConflict, finalUser)
}
