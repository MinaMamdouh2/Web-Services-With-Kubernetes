// Package page provides support for query paging.
package page

import (
	"net/http"
	"strconv"

	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/sys/validate"
)

// Response is what is returned when a query call is performed.
type Response[T any] struct {
	Items       []T `json:"items"`
	Total       int `json:"total"`
	Page        int `json:"page"`
	RowsPerPage int `json:"rowsPerPage"`
}

// NewResponse constructs a repsonse value for a web response.
func NewResponse[T any](items []T, total int, page int, rowsPerPage int) Response[T] {
	return Response[T]{
		Items:       items,
		Total:       total,
		Page:        page,
		RowsPerPage: rowsPerPage,
	}
}

// Page represents the requested page and rows per page.
type Page struct {
	Number      int
	RowsPerPage int
}

// Parse parses the request for the page and rows query string. The
// defaults are provided as well.
func Parse(r *http.Request) (Page, error) {
	values := r.URL.Query()

	number := 1
	if page := values.Get("page"); page != "" {
		var err error
		number, err = strconv.Atoi(page)
		if err != nil {
			return Page{}, validate.NewFieldsError("page", err)
		}
	}

	rowsPerPage := 10
	if rows := values.Get("rows"); rows != "" {
		var err error
		rowsPerPage, err = strconv.Atoi(rows)
		if err != nil {
			return Page{}, validate.NewFieldsError("rows", err)
		}
	}

	return Page{
		Number:      number,
		RowsPerPage: rowsPerPage,
	}, nil
}
