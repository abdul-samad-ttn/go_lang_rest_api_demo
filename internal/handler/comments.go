package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/abdul-samad-ttn/go_lang_rest_api_demo.git/constants"
	"github.com/abdul-samad-ttn/go_lang_rest_api_demo.git/internal/models"
	"github.com/abdul-samad-ttn/go_lang_rest_api_demo.git/internal/store"
)

type CommentsHandler struct {
	commentStore *store.CommentStore
}

// NewCommentsHandler is a constructor for the CommentsHandler struct
func NewCommentsHandler(commentStore *store.CommentStore) *CommentsHandler {
	return &CommentsHandler{commentStore: commentStore}
}

// ListComments is a handler for the GET /comments endpoint
// It lists all comments for a given post id
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

	// list the comments for the given post id
	comments, commentErr := c.commentStore.ListComments(numPostId)

	if commentErr != nil {
		// if there is an error, return an internal server error
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

// GetCommentById is a handler for the GET /comments/:commentId endpoint
// It gets a comment by its id
func (c *CommentsHandler) GetCommentById(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

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

// CreateComment is a handler for the POST /comments endpoint
// It creates a new comment for a given post id
func (c *CommentsHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	postId := r.PathValue("postId")
	if postId == "" {
		http.Error(w, constants.InvalidPostIdErrorMessage, http.StatusBadRequest)
		return
	}

	numPostId, strConvErr := strconv.ParseInt(postId, 10, 64)
	if strConvErr != nil || numPostId == 0 {
		http.Error(w, constants.InvalidPostIdErrorMessage, http.StatusBadRequest)
		return
	}

	var createCommentRequest models.CreateCommentRequest

	errDecode := json.NewDecoder(r.Body).Decode(&createCommentRequest)

	if errDecode != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if createCommentRequest.Name == "" {
		http.Error(w, constants.InvalidNameErrorMessage, http.StatusBadRequest)
		return
	}

	if createCommentRequest.Email == "" {
		http.Error(w, constants.InvalidEmailErrorMessage, http.StatusBadRequest)
		return
	}

	if createCommentRequest.Body == "" {
		http.Error(w, constants.InvalidBodyErrorMessage, http.StatusBadRequest)
		return
	}

	if createCommentRequest.UserId == 0 {
		http.Error(w, constants.InvalidUserIdErrorMessage, http.StatusBadRequest)
		return
	}

	newComment, createCommentErr := c.commentStore.CreateComment(numPostId, createCommentRequest)

	if createCommentErr != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	jsonErr := json.NewEncoder(w).Encode(newComment)

	if jsonErr != nil {
		http.Error(w, constants.InternalServerErrorMessage, http.StatusInternalServerError)
		return
	}
}

func (c *CommentsHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	commentId := r.PathValue("commentId")
	userId := r.URL.Query().Get("userId")

	if commentId == "" {
		http.Error(w, constants.InvalidCommentIdErrorMessage, http.StatusBadRequest)
		return
	}

	if userId == "" {
		http.Error(w, constants.InvalidUserIdErrorMessage, http.StatusBadRequest)
		return
	}

	numCommentId, strConvErr := strconv.ParseInt(commentId, 10, 64)

	if strConvErr != nil || numCommentId == 0 {
		http.Error(w, constants.InvalidCommentIdErrorMessage, http.StatusBadRequest)
		return
	}

	numUserId, strConvErr := strconv.ParseInt(userId, 10, 64)

	if strConvErr != nil || numUserId == 0 {
		http.Error(w, constants.InvalidUserIdErrorMessage, http.StatusBadRequest)
		return
	}

	deleteCommentErr := c.commentStore.DeleteComment(numCommentId, numUserId)

	if deleteCommentErr != nil {
		if errors.Is(deleteCommentErr, sql.ErrNoRows) {
			http.Error(w, constants.CommentNotFoundErrorMessage, http.StatusNotFound)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *CommentsHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	// get the comment id from the url path params
	commentId := r.PathValue("commentId")

	if commentId == "" {
		http.Error(w, constants.InvalidCommentIdErrorMessage, http.StatusBadRequest)
		return
	}

	numCommentId, strConvErr := strconv.ParseInt(commentId, 10, 64)
	if strConvErr != nil || numCommentId == 0 {
		http.Error(w, constants.InvalidCommentIdErrorMessage, http.StatusBadRequest)
		return
	}

	// get the user id from the url query search params
	userId := r.URL.Query().Get("userId")
	if userId == "" {
		http.Error(w, constants.InvalidUserIdErrorMessage, http.StatusBadRequest)
		return
	}

	numUserId, strConvErr := strconv.ParseInt(userId, 10, 64)
	if strConvErr != nil || numUserId == 0 {
		http.Error(w, constants.InvalidUserIdErrorMessage, http.StatusBadRequest)
		return
	}

	existingCommentAuthorUserId, existingCommentAuthorUserIdErr := c.commentStore.GetCommentAuthorUserId(numCommentId)

	if existingCommentAuthorUserIdErr != nil {
		if errors.Is(existingCommentAuthorUserIdErr, sql.ErrNoRows) {
			http.Error(w, constants.CommentNotFoundErrorMessage, http.StatusNotFound)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if existingCommentAuthorUserId != numUserId {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	existingComment, existingCommentErr := c.commentStore.GetCommentById(numCommentId)

	if existingCommentErr != nil {
		if errors.Is(existingCommentErr, sql.ErrNoRows) {
			http.Error(w, constants.CommentNotFoundErrorMessage, http.StatusNotFound)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var updateCommentRequest models.UpdateCommentRequest

	// decode the request body from json into the updateCommentRequest struct
	errDecode := json.NewDecoder(r.Body).Decode(&updateCommentRequest)

	if errDecode != nil {
		if errors.Is(errDecode, io.EOF) {
			http.Error(w, constants.EmptyBodyErrorMessage, http.StatusBadRequest)
			return
		}
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if updateCommentRequest.Name == nil && updateCommentRequest.Body == nil {
		http.Error(w, constants.EmptyBodyErrorMessage, http.StatusBadRequest)
		return
	}

	// initialize the name and body variables with the existing comment values
	var name, body string = existingComment.Name, existingComment.Body

	// if the name is provided in the request body, update the name variable
	if updateCommentRequest.Name != nil {
		if *updateCommentRequest.Name == "" {
			http.Error(w, constants.InvalidNameErrorMessage, http.StatusBadRequest)
			return
		}
		name = *updateCommentRequest.Name
	}

	// if the body is provided in the request body, update the body variable
	if updateCommentRequest.Body != nil {
		if *updateCommentRequest.Body == "" {
			http.Error(w, constants.InvalidBodyErrorMessage, http.StatusBadRequest)
			return
		}
		body = *updateCommentRequest.Body
	}

	// update the comment in the database
	updatedComment, updateCommentErr := c.commentStore.UpdateComment(numCommentId, numUserId, name, body)

	if updateCommentErr != nil {
		if errors.Is(updateCommentErr, sql.ErrNoRows) {
			http.Error(w, constants.CommentNotFoundErrorMessage, http.StatusNotFound)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// return the updated comment in the response body
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonErr := json.NewEncoder(w).Encode(updatedComment)

	if jsonErr != nil {
		http.Error(w, constants.InternalServerErrorMessage, http.StatusInternalServerError)
		return
	}

}
