package api

import (
	"encoding/json"
	"net/http"
)

const (
	CodeUnauthorised = 1001

	CodeMalformedRequest = 2000
	CodeInvalidRequest   = 2001
	CodeNotFound         = 2002
	CodeDuplicatedEntry  = 2003
	CodeForbidden        = 2004

	CodeInternalError = 3001
)

type Response struct {
	Body   interface{}
	Status int
}

func (r Response) Write(w http.ResponseWriter) error {
	w.WriteHeader(r.Status)
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(r.Body)
}

type Error struct {
	Error string `json:"error,omitempty"`
	Code  int    `json:"code"`
}

func OK(w http.ResponseWriter, body interface{}) error {
	return Response{
		Body:   body,
		Status: http.StatusOK,
	}.Write(w)
}

func Created(w http.ResponseWriter, body interface{}) error {
	return Response{
		Body:   body,
		Status: http.StatusCreated,
	}.Write(w)
}

func InternalError(w http.ResponseWriter) error {
	return Response{
		Body: Error{
			Code:  CodeInternalError,
			Error: "internal error",
		},
		Status: http.StatusInternalServerError,
	}.Write(w)
}

func Unauthorised(w http.ResponseWriter) error {
	return Response{
		Body: Error{
			Code:  CodeUnauthorised,
			Error: "unauthorised request",
		},
		Status: http.StatusUnauthorized,
	}.Write(w)
}

func Forbidden(w http.ResponseWriter) error {
	return Response{
		Body: Error{
			Code:  CodeForbidden,
			Error: "forbidden request",
		},
		Status: http.StatusForbidden,
	}.Write(w)
}

func BadRequest(w http.ResponseWriter, code int, text string) error {
	return Response{
		Body: Error{
			Error: text,
			Code:  code,
		},
		Status: http.StatusBadRequest,
	}.Write(w)
}

func NotFound(w http.ResponseWriter, text string) error {
	return Response{
		Body: Error{
			Error: text,
			Code:  CodeNotFound,
		},
		Status: http.StatusNotFound,
	}.Write(w)
}
