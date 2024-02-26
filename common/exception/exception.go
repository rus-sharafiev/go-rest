package exception

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Error struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

type ErrorWithFields struct {
	StatusCode int               `json:"statusCode"`
	Message    string            `json:"message"`
	Fields     map[string]string `json:"fields"`
}

func getError(err error) string {
	parts := strings.Split(err.Error(), ": ")
	if len(parts) > 1 {
		part := parts[1]
		return cases.Upper(language.English).String(part[:1]) + part[1:]
	}
	return parts[0]
}

func MethodNotAllowed(w http.ResponseWriter) {
	err := Error{
		StatusCode: http.StatusMethodNotAllowed,
		Message:    http.StatusText(http.StatusMethodNotAllowed),
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
	json.NewEncoder(w).Encode(&err)
}

func BadRequest(w http.ResponseWriter) {
	err := Error{
		StatusCode: http.StatusBadRequest,
		Message:    http.StatusText(http.StatusBadRequest),
	}
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(&err)
}

func BadRequestError(w http.ResponseWriter, e error) {
	err := Error{
		StatusCode: http.StatusBadRequest,
		Message:    getError(e),
	}
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(&err)
}

func BadRequestMessage(w http.ResponseWriter, msg string) {
	err := Error{
		StatusCode: http.StatusBadRequest,
		Message:    msg,
	}
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(&err)
}

func BadRequestFields(w http.ResponseWriter, fields map[string]string) {
	err := ErrorWithFields{
		StatusCode: http.StatusBadRequest,
		Message:    http.StatusText(http.StatusBadRequest),
		Fields:     fields,
	}
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(&err)
}

func Forbidden(w http.ResponseWriter) {
	err := Error{
		StatusCode: http.StatusForbidden,
		Message:    "You are not allowed to access the requested resource",
	}
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(&err)
}

func NotFound(w http.ResponseWriter) {
	err := Error{
		StatusCode: http.StatusNotFound,
		Message:    http.StatusText(http.StatusNotFound),
	}
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(&err)
}

func Unauthorized(w http.ResponseWriter) {
	err := Error{
		StatusCode: http.StatusUnauthorized,
		Message:    http.StatusText(http.StatusUnauthorized),
	}
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(&err)
}

func UnauthorizedError(w http.ResponseWriter, e error) {
	err := Error{
		StatusCode: http.StatusUnauthorized,
		Message:    getError(e),
	}
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(&err)
}

func InternalServerError(w http.ResponseWriter, e error) {
	err := Error{
		StatusCode: http.StatusInternalServerError,
		Message:    getError(e),
	}
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(&err)
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func PgxNoRows(w http.ResponseWriter, err error) {
	if err == pgx.ErrNoRows {
		NotFound(w)
	} else {
		InternalServerError(w, err)
	}
}
