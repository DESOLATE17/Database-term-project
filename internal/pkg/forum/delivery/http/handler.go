package handler

import (
	"encoding/json"
	"github.com/DESOLATE17/Database-term-project/internal/models"
	"github.com/DESOLATE17/Database-term-project/internal/pkg/forum"
	"github.com/DESOLATE17/Database-term-project/internal/utils"
	"github.com/gorilla/mux"
	"github.com/mailru/easyjson"
	"net/http"
	"strconv"
	"strings"
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
	_ = easyjson.UnmarshalFromReader(r.Body, &user)
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
	_ = easyjson.UnmarshalFromReader(r.Body, &user)
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
	_ = easyjson.UnmarshalFromReader(r.Body, &forum)

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
	_ = decoder.Decode(&posts)

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
		utils.Response(w, http.StatusConflict, models.ErrorResponse{Message: "Parent post was created in another thread"})
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
	_ = easyjson.UnmarshalFromReader(r.Body, &thread)
	thread.Forum = slug

	thread, err := h.uc.CreateForumThread(r.Context(), thread)
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

func (h *Handler) ThreadInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slugOrId, found := vars["slug_or_id"]
	if !found {
		utils.Response(w, http.StatusNotFound, nil)
		return
	}
	finalThread, err := h.uc.CheckThreadIdOrSlug(r.Context(), slugOrId)
	if err == models.NotFound {
		utils.Response(w, http.StatusNotFound, slugOrId)
		return
	}
	utils.Response(w, http.StatusOK, finalThread)
}

func (h *Handler) GetPostsOfThread(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slugOrId, found := vars["slug_or_id"]
	if !found {
		utils.Response(w, http.StatusNotFound, nil)
		return
	}
	params := models.SortParams{}
	params.Desc = r.URL.Query().Get("desc")
	params.Limit = r.URL.Query().Get("limit")
	params.Since = r.URL.Query().Get("since")
	params.Sort = r.URL.Query().Get("sort")

	thread, err := h.uc.CheckThreadIdOrSlug(r.Context(), slugOrId)
	if err == models.NotFound {
		utils.Response(w, http.StatusNotFound, slugOrId)
		return
	}

	finalPosts, err := h.uc.GetPostOfThread(r.Context(), params, thread.ID)
	if err == nil {
		utils.Response(w, http.StatusOK, finalPosts)
		return
	}
	utils.Response(w, http.StatusNotFound, slugOrId)
}

func (h *Handler) GetForumThreads(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug, found := vars["slug"]
	if !found {
		utils.Response(w, http.StatusNotFound, nil)
		return
	}
	params := models.SortParams{}
	params.Desc = r.URL.Query().Get("desc")
	params.Limit = r.URL.Query().Get("limit")
	params.Since = r.URL.Query().Get("since")

	forumS := models.Forum{Slug: slug}

	threads, err := h.uc.GetForumThreads(r.Context(), forumS, params)
	if err == nil {
		utils.Response(w, http.StatusOK, threads)
		return
	}
	utils.Response(w, http.StatusNotFound, slug)
}

func (h *Handler) Vote(w http.ResponseWriter, r *http.Request) {
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

	vote := models.Vote{}
	_ = easyjson.UnmarshalFromReader(r.Body, &vote)

	if thread.ID != 0 {
		vote.Thread = thread.ID
	}

	err = h.uc.Vote(r.Context(), vote)
	if err != nil {
		utils.Response(w, http.StatusNotFound, slugOrId)
		return
	}

	finalThread, _ := h.uc.CheckThreadIdOrSlug(r.Context(), slugOrId)
	utils.Response(w, http.StatusOK, finalThread)
}

func (h *Handler) UpdateThreadInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slugOrId, found := vars["slug_or_id"]
	if !found {
		utils.Response(w, http.StatusNotFound, nil)
		return
	}
	thread := models.Thread{}
	_ = easyjson.UnmarshalFromReader(r.Body, &thread)
	finalThread, err := h.uc.UpdateThreadInfo(r.Context(), slugOrId, thread)
	if err == nil {
		utils.Response(w, http.StatusOK, finalThread)
		return
	}
	utils.Response(w, http.StatusNotFound, slugOrId)
}

func (h *Handler) GetUsersOfForum(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug, found := vars["slug"]
	if !found {
		utils.Response(w, http.StatusNotFound, nil)
		return
	}
	params := models.SortParams{}
	params.Desc = r.URL.Query().Get("desc")
	params.Limit = r.URL.Query().Get("limit")
	params.Since = r.URL.Query().Get("since")

	if params.Limit == "" {
		params.Limit = "100"
	}

	forum := models.Forum{Slug: slug}

	users, err := h.uc.GetUsersOfForum(r.Context(), forum, params)
	if err == nil {
		utils.Response(w, http.StatusOK, users)
		return
	}
	utils.Response(w, http.StatusNotFound, slug)
}

func (h *Handler) GetPostInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idV, found := vars["id"]
	if !found {
		utils.Response(w, http.StatusNotFound, nil)
		return
	}

	id, _ := strconv.Atoi(idV)
	query := r.URL.Query()

	var related []string
	if relateds := query["related"]; len(relateds) > 0 {
		related = strings.Split(relateds[0], ",")
	}

	postFull := models.PostFull{}

	postFull.Post.ID = id
	finalPost, err := h.uc.GetFullPostInfo(r.Context(), postFull, related)
	if err == nil {
		utils.Response(w, http.StatusOK, finalPost)
		return
	}
	utils.Response(w, http.StatusNotFound, id)
}

func (h *Handler) UpdatePostInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ids, found := vars["id"]
	if !found {
		utils.Response(w, http.StatusNotFound, nil)
		return
	}

	postUpdate := models.PostUpdate{}
	_ = easyjson.UnmarshalFromReader(r.Body, &postUpdate)

	id, err := strconv.Atoi(ids)

	if err == nil {
		postUpdate.ID = id
	}

	finalPost, err := h.uc.UpdatePostInfo(r.Context(), postUpdate)
	if err == nil {
		utils.Response(w, http.StatusOK, finalPost)
		return
	}
	utils.Response(w, http.StatusNotFound, id)
}

func (h *Handler) GetClear(w http.ResponseWriter, r *http.Request) {
	h.uc.GetClear(r.Context())
	utils.Response(w, http.StatusOK, nil)
}

func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status := h.uc.GetStatus(r.Context())
	utils.Response(w, http.StatusOK, status)
}
