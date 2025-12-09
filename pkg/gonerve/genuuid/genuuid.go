package genuuid

import "github.com/google/uuid"

type GeneratorUUID interface {
	V4() string
}

type generator struct{}

func New() GeneratorUUID {
	return generator{}
}

func (generator) V4() string {
	return uuid.New().String()
}
