package module

import "github.com/google/uuid"

type ModuleID uuid.UUID

func (id ModuleID) String() string {
	return uuid.UUID(id).String()
}
