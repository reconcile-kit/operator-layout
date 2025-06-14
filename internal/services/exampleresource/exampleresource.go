package exampleresource

type ExampleResourceRepository interface {
}

type Service struct {
	exampleResourceRepository ExampleResourceRepository
}

func NewService(exampleResourceRepository ExampleResourceRepository) *Service {
	return &Service{exampleResourceRepository: exampleResourceRepository}
}
