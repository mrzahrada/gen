package commands

import "context"

type Service struct{}

func New() (*Service, error) {
	return nil, nil
}

func (svc *Service) Command1(ctx context.Context, input Commnand1Input) (*Commnand1Output, error) {
	return nil, nil
}

func (svc Service) Command2(ctx context.Context, input Commnand1Input) (*Commnand1Output, error) {
	return nil, nil
}

type Commnand1Input struct {
	hello string
}

type Commnand1Output struct {
	hello string
}
