package admin

import (
	"net/mail"

	"github.com/google/uuid"
)

type AdminID uuid.UUID

func (id AdminID) String() string {
	return uuid.UUID(id).String()
}

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
