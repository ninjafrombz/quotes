// Filename: cmd/api/token.go

package main

import (
	"errors"
	"net/http"
	"time"

	"quotesapi.desireamagwula.net/internals/data"
	"quotesapi.desireamagwula.net/internals/validator"
)

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Parse rthe email and the password from the request table 
	var input struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// Validate the email and password 
	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Get user details based on the provided email 
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Check if the passwrd matches
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// If passwords don't match, then return an invalid credentials response 
	if !match {
		app.invalidCredentialResponse(w, r)
		return
	}
	// Password is correct so we will generate an authentication token
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Return 
	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}


}