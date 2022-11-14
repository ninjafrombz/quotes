// Filename

package main 

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
)


func (app *application) routes() http.Handler {
	// Create
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/v1/schools", app.listSchoolsHandler)
	router.HandlerFunc(http.MethodPost, "/v1/schools", app.createSchoolHandler)
	router.HandlerFunc(http.MethodGet, "/v1/schools/:id", app.showSchoolHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/schools/:id", app.updateSchoolHandler)
    router.HandlerFunc(http.MethodDelete, "/v1/schools/:id", app.deleteSchoolHandler)
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	return app.recoverPanic(app.rateLimit(router))
}