// Filename: /internals/data/quotes.go

package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"quotesapi.desireamagwula.net/internals/validator"
	"github.com/lib/pq"
)

type Quote struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Author      string    `json:"author"`
	Quote_string     string    `json:"quote_string"`
	Category      []string  `json:"category"`
	Version   int32     `json:"version"`
}

func ValidateQuote(v *validator.Validator, quote *Quote) {
	// Use the Check() Method to execute our validation checks
	v.Check(quote.Author != "", "author", "must be provided")
	v.Check(len(quote.Author) <= 200, "author", "must not be more than 200 bytes long")

	v.Check(quote.Quote_string != "", "quote_string", "must be provided")
	v.Check(len(quote.Quote_string) <= 200, "quote_string", "must not be more than 200 bytes long")


	v.Check(quote.Category != nil, "category", "must be provided!")
	v.Check(len(quote.Category) >= 1, "category", "must contain at least one entry")
	v.Check(len(quote.Category) <= 5, "category", "must contain at most five entries")
	v.Check(validator.Unique(quote.Category), "category", "must not contain duplicate entries")

}

type QuoteModel struct {
	DB *sql.DB
}

// Insert() allows us to create a new quote

func (m QuoteModel) Insert(quote *Quote) error {
	query := `
		INSERT INTO quotes (author, quote_string, category)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, version
	`
	// Collect the data fields into a slice
	args := []interface{}{
		quote.Author, quote.Quote_string,
		pq.Array(quote.Category),
	}
	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&quote.ID, &quote.CreatedAt, &quote.Version)
	//return m.DB.QueryRow(query, args...).Scan(&quote.ID, &quote.CreatedAt, &quote.Version)
}

// Get() allows us to retrieve

func (m QuoteModel) Get(id int64) (*Quote, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// Create the query
	query := `
		SELECT id, created_at, author, quote_string, category, version
		FROM quotes
		WHERE id = $1
	`
	// Declare a quote variable to hold the returned data
	var quote Quote
	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	// Execute the query using QueryRow()
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&quote.ID,
		&quote.CreatedAt,
		&quote.Author,
		&quote.Quote_string,
		pq.Array(&quote.Category),
		&quote.Version,
	)
	// Handle any errors
	if err != nil {
		// Check the type of error
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// Success
	return &quote, nil
}

// Update() allows us to edit/alter a specific quote

func (m QuoteModel) Update(quote *Quote) error {
	// Create a query
	query := `
		UPDATE quotes
		SET author = $1, quote_string = $2,
		category = $3, version = version + 1
		WHERE id = $4
		AND version = $5
		RETURNING version
	`
	args := []interface{}{
		quote.Author,
		quote.Quote_string,
		pq.Array(quote.Category),
		quote.ID,
		quote.Version,
	}

	//Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	//Cleanup to prevent memory leaks
	defer cancel()
	// Check for edit conflicts
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&quote.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil

}

// Delete removes a specific quote
func (m QuoteModel) Delete(id int64) error {

	if id < 1 {
		return ErrRecordNotFound
	}
	// Create the delete query
	query := `
		DELETE FROM quotes
		WHERE id = $1
	`
	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	// Execute the query
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	// Check how many rows were affected by the delete operation. We
	// call the RowsAffected() method on the result variable
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	// Check if no rows were affected
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil

}
func (m QuoteModel) GetAll(author string, quote_string string, category []string, filters Filters) ([]*Quote, Metadata, error) {
	// Construct the query
	query := fmt.Sprintf(`
		SELECT COUNT (*) OVER(), id, created_at, author, quote_string,
			   category, version
		FROM quotes
		WHERE (to_tsvector('simple', author) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (to_tsvector('simple', quote_string) @@ plainto_tsquery('simple', $2) OR $2 = '')
		AND (category @> $3 OR $3 = '{}' )
		ORDER by %s %s, id ASC
		LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortOrder())
	// Create
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	args := []interface{}{author, quote_string, pq.Array(category), filters.limit(), filters.offSet()}
	// Execute the query
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()
	totalRecords := 0
	// Initialize an empty slice
	quotes := []*Quote{}
	// iterate over the rows in the resultset
	for rows.Next() {
		var quote Quote
		// SCan the valuies from the row into the quote
		err := rows.Scan(
			&totalRecords,
			&quote.ID,
			&quote.CreatedAt,
			&quote.Author,
			&quote.Quote_string,
			pq.Array(&quote.Category),
			&quote.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		quotes = append(quotes, &quote)

	}
	// Check if any errors occured after looping through the resultset
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	// safely return the resultset
	return quotes, metadata, nil
}
