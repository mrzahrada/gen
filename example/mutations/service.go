package mutations

import (
	"context"

	"github.com/mrzahrada/gen/example/events"
)

type Mutation struct{}

func New() (Mutation, error) {
	return Mutation{}, nil
}

func (svc Mutation) OnEvent1(ctx context.Context, input *events.Event1) error {
	return nil
}

func (svc Mutation) Push(ctx context.Context) error {
	return nil
}
