package user

import (
	"net/mail"
)

type Email string

func NewEmail(e string) (Email, error) {
	_, err := mail.ParseAddress(e)
	if err != nil {
		return "", ErrInvalidEmail
	}
	return Email(e), nil
}

func (e Email) String() string {
	return string(e)
}
