package api

import "github.com/philipsahli/innominatus-graph/pkg/storage"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	repository storage.RepositoryInterface
}

func NewResolver(repository storage.RepositoryInterface) *Resolver {
	return &Resolver{
		repository: repository,
	}
}
