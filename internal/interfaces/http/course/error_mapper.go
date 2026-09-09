package coursehttp

import (
	"errors"
	"net/http"

	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"
	"github.com/Jaran-Dev/backend-core/internal/infrastructure/repository"
	rootHttp "github.com/Jaran-Dev/backend-core/internal/interfaces/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeCourseNotFound               = "COURSE_NOT_FOUND"
	CodeTranslationNotFound          = "TRANSLATION_NOT_FOUND"
	CodeInvalidState                 = "INVALID_STATE_TRANSITION"
	CodeMissingPrerequisites         = "MISSING_PREREQUISITES"
	CodePrimaryCategoryNotAssigned   = "PRIMARY_CATEGORY_NOT_ASSIGNED"
	CodePrimaryCategoryAlreadyExists = "PRIMARY_CATEGORY_ALREADY_EXISTS"
	CodeForeignKeyViolated           = "FOREIGN_KEY_VIOLATED"
	CodeConflict                     = "CONFLICT"
)

type CourseErrorMapper struct{}

func NewCourseErrorMapper() *CourseErrorMapper {
	return &CourseErrorMapper{}
}

func (m *CourseErrorMapper) MapError(c *gin.Context, err error) bool {
	switch {
	case errors.Is(err, domainCourse.ErrCourseNotFound):
		c.JSON(http.StatusNotFound, rootHttp.APIErrorResponse{
			Code:    CodeCourseNotFound,
			Message: "The requested course could not be found.",
		})
		return true

	case errors.Is(err, domainCourse.ErrTranslationNotFound):
		c.JSON(http.StatusNotFound, rootHttp.APIErrorResponse{
			Code:    CodeTranslationNotFound,
			Message: "The requested translation could not be found.",
		})
		return true

	case errors.Is(err, domainCourse.ErrInvalidStateTransition):
		c.JSON(http.StatusConflict, rootHttp.APIErrorResponse{
			Code:    CodeInvalidState,
			Message: "The course cannot transition to the requested state.",
		})
		return true

	case errors.Is(err, domainCourse.ErrMissingEnglishTranslation), errors.Is(err, domainCourse.ErrMissingModules):
		c.JSON(http.StatusBadRequest, rootHttp.APIErrorResponse{
			Code:    CodeMissingPrerequisites,
			Message: err.Error(),
		})
		return true

	case errors.Is(err, domainCourse.ErrPrimaryCategoryNotAssigned):
		c.JSON(http.StatusBadRequest, rootHttp.APIErrorResponse{
			Code:    CodePrimaryCategoryNotAssigned,
			Message: "Course must have a primary category assigned to be published.",
		})
		return true

	case errors.Is(err, domainCourse.ErrPrimaryCategoryNotPublished):
		c.JSON(http.StatusBadRequest, rootHttp.APIErrorResponse{
			Code:    CodeMissingPrerequisites,
			Message: "Course primary category must be published.",
		})
		return true

	case errors.Is(err, domainCourse.ErrPrimaryCategoryAlreadyExists):
		c.JSON(http.StatusConflict, rootHttp.APIErrorResponse{
			Code:    CodePrimaryCategoryAlreadyExists,
			Message: "Course already has a primary category assigned.",
		})
		return true

	case errors.Is(err, repository.ErrForeignKeyViolated):
		c.JSON(http.StatusBadRequest, rootHttp.APIErrorResponse{
			Code:    CodeForeignKeyViolated,
			Message: "Referenced course or category does not exist.",
		})
		return true

	case errors.Is(err, domainCourse.ErrInvalidSlug), errors.Is(err, domainCourse.ErrInvalidLocale):
		c.JSON(http.StatusBadRequest, rootHttp.APIErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: err.Error(),
		})
		return true

	case errors.Is(err, domainCourse.ErrConcurrentModification):
		c.JSON(http.StatusConflict, rootHttp.APIErrorResponse{
			Code:    CodeConflict,
			Message: "The course was modified by another transaction. Please refresh and try again.",
		})
		return true
	}

	return false
}
