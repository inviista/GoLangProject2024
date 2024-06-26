package main

import (
	"errors"
	"net/http"

	"api/pkg/library-app/models"
	"api/pkg/library-app/validator"
)

func (app *application) GetBooks(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title         string
		Author        string
		PublishedYear int
		models.Filters
	}

	v := validator.New()
	qs := r.URL.Query()
	// Use our helpers to extract the title, author, publishedyear value range query string values, falling back to the
	// defaults of an empty string and an empty slice, respectively, if they are not provided
	// by the client.
	input.Title = app.readString(qs, "title", "")
	input.Author = app.readString(qs, "author", "")
	input.PublishedYear = app.readInt(qs, "publishedyear", 1, v)
	// Get the page and page_size query string value as integers. Notice that we set the default
	// page value to 1 and default page_size to 20, and that we pass the validator instance
	// as the final argument.
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	// Extract the sort query string value, falling back to "id" if it is not provided
	// by the client (which will imply an ascending sort on menu ID).
	input.Filters.Sort = app.readString(qs, "sort", "id")
	// Add the supported sort values for this endpoint to the sort safelist.
	input.Filters.SortSafelist = []string{"id", "title", "author", "publishedyear", "-id", "-title", "-author", "-publishedyear"}
	// Execute the validation checks on the Filters struct and send a response
	// containing the errors if necessary.
	if models.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Call the GetAll() method to retrieve the movies, passing in the various filter
	// parameters.
	books, metadata, err := app.models.Books.GetAll(input.Title, input.Author, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"books": books, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) GetBook(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	book, err := app.models.Books.Get(id)

	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"book": book}, nil)
}

func (app *application) CreateBook(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title         string `json:"title"`
		Author        string `json:"author"`
		PublishedYear int    `json:"publishedYear"`
	}

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "Invalid request payload")
		return
	}

	book := &models.Book{
		Title:         input.Title,
		Author:        input.Author,
		PublishedYear: input.PublishedYear,
	}

	err = app.models.Books.Insert(book)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusCreated, envelope{"book": book}, nil)
}

func (app *application) DeleteBook(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}
	book, err := app.models.Books.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	app.models.Books.Delete(id)
	app.writeJSON(w, http.StatusOK, envelope{"message": "success", "deleted_book": book}, nil)
}

func (app *application) UpdateBook(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	book, err := app.models.Books.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Title         *string `json:"title"`
		Author        *string `json:"author"`
		PublishedYear *int    `json:"publishedYear"`
	}

	err = app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil {
		book.Title = *input.Title
	}

	if input.Author != nil {
		book.Author = *input.Author
	}

	if input.PublishedYear != nil {
		book.PublishedYear = *input.PublishedYear
	}

	v := validator.New()

	if models.ValidateBook(v, book); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Books.Update(book)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"book": book}, nil)

}
