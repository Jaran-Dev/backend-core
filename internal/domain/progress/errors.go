package progress

import "errors"

var (
	ErrEnrollmentNotFound = errors.New("enrollment not found")
	ErrProgressNotFound   = errors.New("lesson progress not found")
	ErrAlreadyEnrolled    = errors.New("user is already enrolled in this course")
	ErrNotEnrolled        = errors.New("user is not enrolled in this course")
)
