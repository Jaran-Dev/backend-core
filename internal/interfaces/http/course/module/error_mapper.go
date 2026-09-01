package modulehttp

import (
	"errors"
	"net/http"

	domainModule "github.com/Jaran-Dev/backend-core/internal/domain/course/module"
	rootHttp "github.com/Jaran-Dev/backend-core/internal/interfaces/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeModuleNotFound        = "MODULE_NOT_FOUND"
	CodeInvalidPosition       = "INVALID_POSITION"
	CodeInvalidModuleTitle    = "INVALID_MODULE_TITLE"
	CodeInvalidReorderPayload = "INVALID_REORDER_PAYLOAD"
	CodeConflict              = "CONFLICT"
)

type ModuleErrorMapper struct{}

func NewModuleErrorMapper() *ModuleErrorMapper {
	return &ModuleErrorMapper{}
}

func (m *ModuleErrorMapper) MapError(c *gin.Context, err error) bool {
	switch {
	case errors.Is(err, domainModule.ErrModuleNotFound):
		c.JSON(http.StatusNotFound, rootHttp.APIErrorResponse{
			Code:    CodeModuleNotFound,
			Message: "The requested module could not be found.",
		})
		return true

	case errors.Is(err, domainModule.ErrInvalidPosition):
		c.JSON(http.StatusBadRequest, rootHttp.APIErrorResponse{
			Code:    CodeInvalidPosition,
			Message: "Module position must be 0 or greater.",
		})
		return true

	case errors.Is(err, domainModule.ErrInvalidModuleTitle):
		c.JSON(http.StatusBadRequest, rootHttp.APIErrorResponse{
			Code:    CodeInvalidModuleTitle,
			Message: "Module title cannot be empty.",
		})
		return true

	case errors.Is(err, domainModule.ErrInvalidReorderPayload):
		c.JSON(http.StatusBadRequest, rootHttp.APIErrorResponse{
			Code:    CodeInvalidReorderPayload,
			Message: "Reorder failed. The payload must include all module IDs for the course exactly once.",
		})
		return true

	case errors.Is(err, domainModule.ErrConcurrentModification):
		c.JSON(http.StatusConflict, rootHttp.APIErrorResponse{
			Code:    CodeConflict,
			Message: "The module was modified by another transaction. Please refresh and try again.",
		})
		return true
	}

	// Not a module error
	return false
}
