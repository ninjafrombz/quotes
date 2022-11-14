// Filename:
package main

import (
	"errors"
	"fmt"
	"net/http"

	"quotesapi.desireamagwula.net/internals/data"
	"quotesapi.desireamagwula.net/internals/validator"
)

// CreatequoteHandler for the POST /v1/quotes" endpoint

func (app *application) createQuoteHandler(w http.ResponseWriter, r *http.Request) {
	// Our target decode destination fmt.Fprintln(w, "create a new quote..")
	var input struct {
		Author    string   `json:"author"`
		Quote_string   string   `json:"quote_string"`
		Category    []string `json:"categoty"`
	}

	// Initialize a new json.Decoder instance
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Copy the values from the input struct to a new quote struct
	quote := &data.Quote{
		Author:    input.Author,
		Quote_string:   input.Quote_string,
		Category:    input.Category,
	}

	//Initialize a new validator instance
	v := validator.New()

	// Check the map to determine if there were any validation errors
	if data.ValidateQuote(v, quote); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// CReate a quote
	err = app.models.Quote.Insert(quote)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	// CReate a location header for the newly created
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/quotes/%d", quote.ID))
	//Write the JSON response with 201 - Created status code with the body
	// being the quote data and the header being the headers map

	err = app.writeJSON(w, http.StatusCreated, envelope{"quote": quote}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)

	}

}

func (app *application) showQuoteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch the specific quote
	quote, err := app.models.Quote.Get(id)
	// Handle errors
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	// Write the sdata returned by Get()
	err = app.writeJSON(w, http.StatusOK, envelope{"quote": quote}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateQuoteHandler(w http.ResponseWriter, r *http.Request) {
	// This method does a partial replacement
	// Get the id for the quote that needs updating
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Fetch the orginal record from the database
	quote, err := app.models.Quote.Get(id)
	// Handle errors
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Create an input struct to hold data read in from the client
	// We update input struct to use pointers because pointers have a
	// default value of nil
	// If a field remains nil then we know that the client did not update it
	var input struct {
		Author    *string  `json:"author"`
		Quote_string   *string  `json:"quote_string"`
		Category    []string `json:"category"`
	}

	// Initialize a new json.Decoder instance
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// Check for updates
	if input.Author != nil {
		quote.Author = *input.Author
	}
	if input.Quote_string != nil {
		quote.Quote_string = *input.Quote_string
	}
	if input.Category != nil {
		quote.Category = input.Category
	}

	// Perform validation on the updated quote. If validation fails, then
	// we send a 422 - Unprocessable Entity respose to the client
	// Initialize a new Validator instance
	v := validator.New()

	// Check the map to determine if there were any validation errors
	if data.ValidateQuote(v, quote); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Let's pass the updated quote record to the Update() method
	err = app.models.Quote.Update(quote)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Write the data returned by Get()
	err = app.writeJSON(w, http.StatusOK, envelope{"quote": quote}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteQuoteHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Delete the quote from the Database. Send a 404 not found status cide to the client
	// if not found

	err = app.models.Quote.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}
	// Return 200 Status OK to the client with a success message
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "quote successfuly deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// Allows the client to see a listing of quotes based on a set of criterias

func (app *application) listQuotesHandler(w http.ResponseWriter, r *http.Request) {
	// Create an input struct to hold our query parameters
	var input struct {
		Author  string
		Quote_string string
		Category  []string
		data.Filters
	}
	v := validator.New()
	// Get the url values map
	qs := r.URL.Query()
	// Use the helper methods to extfract the values
	input.Author = app.readString(qs, "author", "")
	input.Quote_string = app.readString(qs, "quote_string", "")
	input.Category = app.readCSV(qs, "category", []string{})
	//Get the page information
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	// Get the sort info
	input.Filters.Sort = app.readString(qs, "sort", "id")
	// Specify the allowed sort values
	input.Filters.SortList = []string{"id", "author", "quote_string", "-id", "-author", "-quote_string"}
	// CHeck for validation error
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Get a listing of all quotes
	quotes, metadata, err := app.models.Quote.GetAll(input.Author, input.Quote_string, input.Category, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Send a JSON response containing all the quotes
	err = app.writeJSON(w, http.StatusOK, envelope{"quotes": quotes, "metadata ": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}
