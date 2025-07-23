package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"

	"github.com/hashicorp/go-multierror"
	"github.com/vishenosik/gocherry/pkg/errors"
)

// ErrorResponse represents an http API error
type ErrorResponse struct {
	Message string   `json:"message,omitempty"`
	Errors  []string `json:"errors,omitempty"`
}

type httpError struct {
	statusCode int
	message    string
	_error     error
}

func NewError(statusCode int, err error) error {
	if err == nil {
		return nil
	}
	return &httpError{
		statusCode: statusCode,
		message:    http.StatusText(statusCode),
		_error:     err,
	}
}

func (h *httpError) Error() string {
	return fmt.Sprintf("%s [%d]: %s", h.message, h.statusCode, h._error.Error())
}

func (h *httpError) MarshalJSON() ([]byte, error) {
	switch err := h._error.(type) {

	case *multierror.Error:
		errs := make([]string, 0, len(err.Errors))

		for _, e := range err.Errors {
			if e == nil {
				continue
			}
			errs = append(errs, e.Error())
		}
		return json.Marshal(ErrorResponse{
			Message: h.message,
			Errors:  errs,
		})

	case *errors.MultiError:
		return json.Marshal(ErrorResponse{
			Message: h.message,
			Errors:  err.List(),
		})

	default:
		log.Println("default", reflect.TypeOf(h._error).Elem().String())
		return json.Marshal(ErrorResponse{
			Message: h.message,
			Errors:  []string{err.Error()},
		})
	}
}

// sendError sends a JSON error response
func SendErrors(w http.ResponseWriter, statusCode int, _error error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	resp, _ := json.Marshal(_error)
	w.Write(resp)
}

type HandlerWithError func(w http.ResponseWriter, r *http.Request) error

func (h HandlerWithError) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {

		switch e := err.(type) {
		case *httpError:
			SendErrors(w, e.statusCode, err)
			return
		}

		SendErrors(w, http.StatusInternalServerError, err)
	}
}
