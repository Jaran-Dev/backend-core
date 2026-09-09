package profilehttp

import (
	"errors"
	"net/http"

	domainPref "github.com/Jaran-Dev/backend-core/internal/domain/preferences"
	domainUser "github.com/Jaran-Dev/backend-core/internal/domain/user"
	rootHttp "github.com/Jaran-Dev/backend-core/internal/interfaces/http"

	"github.com/gin-gonic/gin"
)

type ProfileErrorMapper struct{}

func NewProfileErrorMapper() *ProfileErrorMapper {
	return &ProfileErrorMapper{}
}

func (m *ProfileErrorMapper) MapError(c *gin.Context, err error) bool {
	switch {
	// User Errors
	case errors.Is(err, domainUser.ErrUserNotFound):
		c.JSON(http.StatusNotFound, rootHttp.APIErrorResponse{Code: "USER_NOT_FOUND", Message: "User profile not found."})
		return true
	case errors.Is(err, domainUser.ErrUsernameTaken):
		c.JSON(http.StatusConflict, rootHttp.APIErrorResponse{Code: "USERNAME_TAKEN", Message: "This username is already taken."})
		return true
	case errors.Is(err, domainUser.ErrInvalidDisplayName), errors.Is(err, domainUser.ErrInvalidUsername),
		errors.Is(err, domainUser.ErrInvalidBio), errors.Is(err, domainUser.ErrInvalidHeadline),
		errors.Is(err, domainUser.ErrInvalidPronouns), errors.Is(err, domainUser.ErrInvalidLocation),
		errors.Is(err, domainUser.ErrInvalidDateOfBirth):
		c.JSON(http.StatusBadRequest, rootHttp.APIErrorResponse{Code: "VALIDATION_ERROR", Message: err.Error()})
		return true

	// Preferences Errors
	case errors.Is(err, domainPref.ErrPreferencesNotFound):
		c.JSON(http.StatusNotFound, rootHttp.APIErrorResponse{Code: "PREFERENCES_NOT_FOUND", Message: "Preferences not found."})
		return true
	case errors.Is(err, domainPref.ErrConcurrentModification):
		c.JSON(http.StatusConflict, rootHttp.APIErrorResponse{Code: "CONFLICT", Message: "Data was modified by another transaction. Please refresh."})
		return true
	case errors.Is(err, domainPref.ErrInvalidLocale), errors.Is(err, domainPref.ErrInvalidTheme),
		errors.Is(err, domainPref.ErrInvalidPlaybackQuality), errors.Is(err, domainPref.ErrInvalidDigestFrequency):
		c.JSON(http.StatusBadRequest, rootHttp.APIErrorResponse{Code: "VALIDATION_ERROR", Message: err.Error()})
		return true
	}

	// Not a profile error
	return false
}
