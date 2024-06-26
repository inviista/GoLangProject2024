package models

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"api/pkg/library-app/validator"
)

type Book struct {
	Id            int    `json:"id"`
	Title         string `json:"title"`
	Author        string `json:"author"`
	PublishedYear int    `json:"publishedYear"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type BookModel struct {
	DB       *sql.DB
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

func (m BookModel) GetAll(title string, author string, filters Filters) ([]*Book, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, title, author, publishedyear, created_at, updated_at
		FROM books
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (to_tsvector('simple', author) @@ plainto_tsquery('simple', $2) OR $2 = '')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// As our SQL query now has quite a few placeholder parameters, let's collect the
	// values for the placeholders in a slice. Notice here how we call the limit() and
	// offset() methods on the Filters struct to get the appropriate values for the
	// LIMIT and OFFSET clauses.

	args := []interface{}{title, author, filters.limit(), filters.offset()}
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()
	totalRecords := 0
	books := []*Book{}

	for rows.Next() {
		var book Book
		err := rows.Scan(
			&totalRecords,
			&book.Id,
			&book.Title,
			&book.Author,
			&book.PublishedYear,
			&book.CreatedAt,
			&book.UpdatedAt,
		)

		if err != nil {
			return nil, Metadata{}, err
		}
		books = append(books, &book)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	// Generate a Metadata struct, passing in the total record count and pagination
	// parameters from the client.
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	// Include the metadata struct when returning.
	return books, metadata, nil
}

func (m BookModel) Get(id int) (*Book, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	query := `
		SELECT id, title, author, publishedYear, created_at, updated_at
		FROM books
		WHERE id = $1
		`

	var book Book

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(&book.Id, &book.Title, &book.Author, &book.PublishedYear, &book.CreatedAt, &book.UpdatedAt)

	if err != nil {
		return nil, ErrRecordNotFound
	}

	return &book, nil
}

func (m BookModel) Insert(book *Book) error {
	query := `
		INSERT INTO books (title, author, publishedYear) 
		VALUES ($1, $2, $3) 
		RETURNING id, created_at, updated_at
	`

	args := []interface{}{book.Title, book.Author, book.PublishedYear}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&book.Id, &book.CreatedAt, &book.UpdatedAt)
}

func (m BookModel) Delete(id int) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `
		DELETE FROM books
		WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, id)
	return err
}

func (m BookModel) Update(book *Book) error {

	query := `
		UPDATE books
		SET title = $1, author = $2, publishedyear = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
		`
	args := []interface{}{book.Title, book.Author, book.PublishedYear, book.Id}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&book.UpdatedAt)
}

func ValidateBook(v *validator.Validator, book *Book) {
	// Check if the title field is empty.
	v.Check(book.Title != "", "title", "must be provided")
	// Check if the title field is not more than 100 characters.
	v.Check(len(book.Title) <= 100, "title", "must not be more than 100 bytes long")
	// Check if the author field is not more than 1000 characters.
	v.Check(len(book.Author) <= 100, "author", "must not be more than 100 bytes long")
}
