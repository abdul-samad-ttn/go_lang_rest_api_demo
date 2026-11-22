package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/abdul-samad-ttn/go_lang_rest_api_demo.git/constants"
	"github.com/abdul-samad-ttn/go_lang_rest_api_demo.git/internal/store"
)

type CommentsHandler struct {
	commentStore *store.CommentStore
}

func NewCommentsHandler(commentStore *store.CommentStore) *CommentsHandler {
	return &CommentsHandler{commentStore: commentStore}
}

func (c *CommentsHandler) ListComments(w http.ResponseWriter, r *http.Request) {

	postId := r.PathValue("postId")
	if postId == "" {
		http.Error(w, constants.InvalidPostIdErrorMessage, http.StatusBadRequest)
		return
	}

	numPostId, strConvErr := strconv.ParseInt(postId, 10, 64)

	if strConvErr != nil {
		http.Error(w, constants.InvalidPostIdErrorMessage, http.StatusBadRequest)
		return
	}

	comments, commentErr := c.commentStore.ListComments(numPostId)

	if commentErr != nil {
		http.Error(w, constants.InternalServerErrorMessage, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	jsonErr := json.NewEncoder(w).Encode(comments)

	if jsonErr != nil {
		http.Error(w, constants.InternalServerErrorMessage, http.StatusInternalServerError)
	}
}

func (c *CommentsHandler) GetCommentById(w http.ResponseWriter, r *http.Request) {

	commentId := r.PathValue("commentId")

	if commentId == "" {
		http.Error(w, constants.InvalidCommentIdErrorMessage, http.StatusBadRequest)
		return
	}

	numCommentId, strConvErr := strconv.ParseInt(commentId, 10, 64)

	if strConvErr != nil {
		http.Error(w, constants.InvalidCommentIdErrorMessage, http.StatusBadRequest)
		return
	}

	comment, commentErr := c.commentStore.GetCommentById(numCommentId)

	if commentErr != nil {
		if errors.Is(commentErr, sql.ErrNoRows) {
			http.Error(w, constants.CommentNotFoundErrorMessage, http.StatusNotFound)
			return
		}
		http.Error(w, constants.InternalServerErrorMessage, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	jsonErr := json.NewEncoder(w).Encode(comment)

	if jsonErr != nil {
		http.Error(w, constants.InternalServerErrorMessage, http.StatusInternalServerError)
		return
	}
}
