// Filename: Internals/data/models.go

package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict = errors.New("edit conflict")
)

type Models struct {
	Quote QuoteModel
	Tokens TokenModel
	Users UserModel
}

// NewModels() allows us to create a new MOdels 

func NewModels(db *sql.DB) Models {
	return Models{
		Quote: QuoteModel{DB: db},
		Tokens: TokenModel{DB: db},
		Users:     UserModel{DB: db},
	}
} 