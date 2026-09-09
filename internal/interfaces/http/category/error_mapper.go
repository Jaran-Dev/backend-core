package categoryhttp

import (
	"errors"
	"net/http"

	domainCategory "github.com/Jaran-Dev/backend-core/internal/domain/category"
	rootHttp "github.com/Jaran-Dev/backend-core/internal/interfaces/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeCategoryNotFound       = "CATEGORY_NOT_FOUND"
	CodeCategoryHasCourses     = "CATEGORY_HAS_COURSES"
	CodeInvalidDepth           = "INVALID_CATEGORY_DEPTH"
	CodeCircularReference      = "CIRCULAR_REFERENCE"
	CodeInvalidStateTransition = "INVALID_STATE_TRANSITION"
	CodeMissingPrerequisites   = "MISSING_PREREQUISITES"
	CodeConflict               = "CONFLICT"
)

type CategoryErrorMapper struct{}

func NewCategoryErrorMapper() *CategoryErrorMapper {
	return &CategoryErrorMapper{}
}

func (m *CategoryErrorMapper) MapError(c *gin.Context, err error) bool {
	switch {
	case errors.Is(err, domainCategory.ErrCategoryNotFound):
		c.JSON(http.StatusNotFound, rootHttp.APIErrorResponse{
			Code:    CodeCategoryNotFound,
			Message: "The requested category could not be found.",
		})
		return true

	case errors.Is(err, domainCategory.ErrCategoryHasCourses):
		c.JSON(http.StatusConflict, rootHttp.APIErrorResponse{
			Code:    CodeCategoryHasCourses,
			Message: "Category cannot be deleted while courses are assigned to it.",
		})
		return true

	case errors.Is(err, domainCategory.ErrInvalidDepth),
		errors.Is(err, domainCategory.ErrMaxDepthExceeded):
		c.JSON(http.StatusBadRequest, rootHttp.APIErrorResponse{
			Code:    CodeInvalidDepth,
			Message: "Category hierarchy max depth is 2.",
		})
		return true

	case errors.Is(err, domainCategory.ErrCircularReference):
		c.JSON(http.StatusBadRequest, rootHttp.APIErrorResponse{
			Code:    CodeCircularReference,
			Message: "Circular reference detected in category hierarchy.",
		})
		return true

	case errors.Is(err, domainCategory.ErrInvalidStateTransition):
		c.JSON(http.StatusConflict, rootHttp.APIErrorResponse{
			Code:    CodeInvalidStateTransition,
			Message: "The category cannot transition to the requested state.",
		})
		return true

	case errors.Is(err, domainCategory.ErrMissingEnglishTranslation):
		c.JSON(http.StatusBadRequest, rootHttp.APIErrorResponse{
			Code:    CodeMissingPrerequisites,
			Message: "Category must have an English translation to be published.",
		})
		return true

	case errors.Is(err, domainCategory.ErrTranslationNotFound):
		c.JSON(http.StatusNotFound, rootHttp.APIErrorResponse{
			Code:    CodeCategoryNotFound,
			Message: "The requested translation could not be found.",
		})
		return true

	case errors.Is(err, domainCategory.ErrConcurrentModification):
		c.JSON(http.StatusConflict, rootHttp.APIErrorResponse{
			Code:    CodeConflict,
			Message: "The category was modified by another transaction. Please refresh and try again.",
		})
		return true

	case errors.Is(err, domainCategory.ErrInvalidSlug), errors.Is(err, domainCategory.ErrInvalidLocale),
		errors.Is(err, domainCategory.ErrInvalidPosition):
		c.JSON(http.StatusBadRequest, rootHttp.APIErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: err.Error(),
		})
		return true
	}

	// Not a category error
	return false
}
