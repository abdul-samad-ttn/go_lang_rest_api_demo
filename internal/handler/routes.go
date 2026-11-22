package handler

import (
	"net/http"

	"github.com/abdul-samad-ttn/go_lang_rest_api_demo.git/constants"
)

func RoutesRegisterHealthCheck(router *http.ServeMux) {
	router.HandleFunc(constants.GET+" "+constants.HealthCheckPath, HealthCheckHandler)
}

func RoutesRegisterPosts(router *http.ServeMux, posts *PostsHandler) {
	router.HandleFunc(constants.GET+" "+constants.PostsPath, posts.ListPosts)
	router.HandleFunc(constants.GET+" "+constants.PostByIdPath, posts.GetPostById)
}

func RoutesRegisterComments(router *http.ServeMux, comments *CommentsHandler) {
	router.HandleFunc(constants.GET+" "+constants.CommentsPath, comments.ListComments)
	router.HandleFunc(constants.GET+" "+constants.CommentByIdPath, comments.GetCommentById)
	router.HandleFunc(constants.POST+" "+constants.CommentsPath, comments.CreateComment)
	router.HandleFunc(constants.DELETE+" "+constants.CommentByIdPath, comments.DeleteComment)
}
