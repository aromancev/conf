package api

import (
	"context"
	"encoding/json"
	"net/http"
)

const (
	CodeUnauthorised = 1001

	CodeMalformedRequest = 2000
	CodeInvalidRequest   = 2001

	CodeInternalError = 3001
)

const (
	traceKey = "traceId"
)

type Response struct {
	Body   interface{}
	Status int
}

func (r Response) Write(ctx context.Context, w http.ResponseWriter) error {
	if t, ok := ctx.Value(traceKey).(string); ok {
		w.Header().Set(traceKey, t)
	}
	w.WriteHeader(r.Status)
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(r.Body)
}

type Error struct {
	Error string `json:"error,omitempty"`
	Code  int    `json:"code"`
}

func OK(body interface{}) Response {
	return Response{
		Body:   body,
		Status: http.StatusOK,
	}
}

func Created(body interface{}) Response {
	return Response{
		Body:   body,
		Status: http.StatusCreated,
	}
}

func InternalError() Response {
	return Response{
		Body: Error{
			Code:  CodeInternalError,
			Error: "internal error",
		},
		Status: http.StatusInternalServerError,
	}
}

func Unauthorised() Response {
	return Response{
		Body: Error{
			Code:  CodeUnauthorised,
			Error: "unauthorised request",
		},
		Status: http.StatusUnauthorized,
	}
}

func BadRequest(code int, text string) Response {
	return Response{
		Body: Error{
			Error: text,
			Code:  code,
		},
		Status: http.StatusBadRequest,
	}
}
