package handler

import (
	"net/http"

	"github.com/abdul-samad-ttn/go_lang_rest_api_demo.git/constants"
)

func RoutesRegister(router *http.ServeMux, posts *PostsHandler) {
	router.HandleFunc(constants.GET+" "+constants.HealthCheckPath, HealthCheckHandler)
	router.HandleFunc(constants.GET+" "+constants.PostsPath, posts.ListPosts)
	router.HandleFunc(constants.GET+" "+constants.PostByIdPath, posts.GetPostById)

}
