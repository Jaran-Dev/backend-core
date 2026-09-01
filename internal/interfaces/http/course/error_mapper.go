package coursehttp

import (
	"errors"
	"net/http"

	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"
	rootHttp "github.com/Jaran-Dev/backend-core/internal/interfaces/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeCourseNotFound       = "COURSE_NOT_FOUND"
	CodeTranslationNotFound  = "TRANSLATION_NOT_FOUND"
	CodeInvalidState         = "INVALID_STATE_TRANSITION"
	CodeMissingPrerequisites = "MISSING_PREREQUISITES"
	CodeConflict             = "CONFLICT"
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

	case errors.Is(err, domainCourse.ErrMissingEnglishTranslation):
		c.JSON(http.StatusBadRequest, rootHttp.APIErrorResponse{
			Code:    CodeMissingPrerequisites,
			Message: "Course must have an English translation to be published.",
		})
		return true

	case errors.Is(err, domainCourse.ErrInvalidSlug), errors.Is(err, domainCourse.ErrInvalidLocale):
		c.JSON(http.StatusBadRequest, rootHttp.APIErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: err.Error(),
		})
		return true
	}

	// Not a course error
	return false
}
