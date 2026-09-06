package user

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidEmail      = errors.New("invalid email format")
	ErrInvalidCredential = errors.New("invalid credentials")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	ErrUsernameTaken     = errors.New("username is already taken")

	// Validation Errors
	ErrInvalidDisplayName = errors.New("invalid display name: must be between 1 and 50 characters")
	ErrInvalidUsername    = errors.New("invalid username: must be 3-30 lowercase alphanumeric characters separated by hyphens")
	ErrInvalidBio         = errors.New("invalid bio: must be 500 characters or less")
	ErrInvalidHeadline    = errors.New("invalid headline: must be 100 characters or less")
	ErrInvalidPronouns    = errors.New("invalid pronouns: must be 20 characters or less")
	ErrInvalidLocation    = errors.New("invalid location: must be 100 characters or less")
	ErrInvalidDateOfBirth = errors.New("invalid date of birth: must be at least 13 years old")
)
