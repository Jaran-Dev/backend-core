package module

import (
	"github.com/google/uuid"
)

type ModuleID uuid.UUID

func (id ModuleID) String() string {
	return uuid.UUID(id).String()
}

type Position int

func NewPosition(p int) (Position, error) {
	if p < 0 {
		return 0, ErrInvalidPosition
	}
	return Position(p), nil
}
