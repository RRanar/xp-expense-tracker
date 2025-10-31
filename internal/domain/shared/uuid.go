package shared

import (
	"github.com/google/uuid"
)

type ID struct {
	value string
}

func NewID() ID {
	return ID{value: uuid.NewString()}
}

func IDFromString(v string) ID {
	return ID{value: v}
}

func (id ID) String() string {
	return id.value
}
