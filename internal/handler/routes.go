package handler

import (
	"net/http"

	"github.com/abdul-samad-ttn/go_lang_rest_api_demo.git/constants"
)

func RoutesRegister(router *http.ServeMux) {
	router.HandleFunc(constants.GET+" "+constants.HealthCheckPath, HealthCheckHandler)

}
