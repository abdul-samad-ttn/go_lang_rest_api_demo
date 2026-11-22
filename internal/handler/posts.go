package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/abdul-samad-ttn/go_lang_rest_api_demo.git/constants"
	"github.com/abdul-samad-ttn/go_lang_rest_api_demo.git/internal/store"
)

type PostsHandler struct {
	postStore *store.PostStore
}

func NewPostsHandler(postStore *store.PostStore) *PostsHandler {
	return &PostsHandler{postStore: postStore}
}

func (p *PostsHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
	posts, postError := p.postStore.ListPosts()

	if postError != nil {
		http.Error(w, constants.InternalServerErrorMessage, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonError := json.NewEncoder(w).Encode(posts)

	if jsonError != nil {
		http.Error(w, constants.InternalServerErrorMessage, http.StatusInternalServerError)
	}
}

func (p *PostsHandler) GetPostById(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")

	numId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, constants.InvalidPostIdErrorMessage, http.StatusBadRequest)
		return
	}

	post, postError := p.postStore.GetPostById(numId)

	if postError != nil {
		http.Error(w, constants.PostNotFoundErrorMessage, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonError := json.NewEncoder(w).Encode(post)

	if jsonError != nil {
		http.Error(w, constants.InternalServerErrorMessage, http.StatusInternalServerError)
		return
	}
}
