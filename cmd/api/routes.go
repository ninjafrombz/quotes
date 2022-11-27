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
	router.HandlerFunc(http.MethodGet, "/v1/Quotes", app.listQuotesHandler)
	router.HandlerFunc(http.MethodPost, "/v1/Quotes", app.createQuoteHandler)
	router.HandlerFunc(http.MethodGet, "/v1/Quotes/:id", app.showQuoteHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/Quotes/:id", app.updateQuoteHandler)
    router.HandlerFunc(http.MethodDelete, "/v1/Quotes/:id", app.deleteQuoteHandler)
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	return app.recoverPanic(app.rateLimit(router))
}