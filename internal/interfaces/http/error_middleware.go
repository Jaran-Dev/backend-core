package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type APIErrorDetail struct {
	Field string `json:"field"`
	Issue string `json:"issue"`
	Value any    `json:"value,omitempty"`
}

type APIErrorResponse struct {
	Code    string           `json:"code"`
	Message string           `json:"message"`
	Details []APIErrorDetail `json:"details,omitempty"`
	TraceID string           `json:"trace_id,omitempty"`
}

type DomainErrorMapper interface {
	MapError(c *gin.Context, err error) bool
}
type ErrorMiddleware struct {
	mappers []DomainErrorMapper
}

func NewErrorMiddleware() *ErrorMiddleware {
	return &ErrorMiddleware{}
}

func (em *ErrorMiddleware) RegisterMapper(m DomainErrorMapper) {
	em.mappers = append(em.mappers, m)
}

type HTTPError struct {
	Status  int
	Code    string
	Message string
}

func (e *HTTPError) Error() string {
	return e.Message
}

func NewHTTPError(status int, code, message string) *HTTPError {
	return &HTTPError{Status: status, Code: code, Message: message}
}

func (em *ErrorMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		// 1. Check for generic HTTP errors (like bad UUIDs in URLs)
		var httpErr *HTTPError
		if errors.As(err, &httpErr) {
			c.JSON(httpErr.Status, APIErrorResponse{
				Code:    httpErr.Code,
				Message: httpErr.Message,
			})
			return
		}

		// 2. Try to map it using registered domain mappers (Course, Module, etc.)
		for _, mapper := range em.mappers {
			if mapper.MapError(c, err) {
				return // A domain mapper handled it and wrote the JSON response
			}
		}

		// 3. If no domain mapper matched, check if it's a standard Gin validation error
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			details := make([]APIErrorDetail, len(valErrs))
			for i, fieldErr := range valErrs {
				details[i] = APIErrorDetail{
					Field: fieldErr.Field(),
					Issue: fieldErr.Tag(),
					Value: fieldErr.Value(),
				}
			}
			c.JSON(http.StatusBadRequest, APIErrorResponse{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid request payload.",
				Details: details,
			})
			return
		}

		// 4. The Catch-All: Unknown internal server error
		slog.Error("Unhandled internal error", "error", err.Error(), "path", c.Request.URL.Path)
		c.JSON(http.StatusInternalServerError, APIErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "An unexpected internal server error occurred.",
		})
	}
}
