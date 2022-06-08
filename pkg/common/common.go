package common

import (
	"context"
	"io"
)

type ControllerInterface interface {
	MakeApiRequest(method, url string, body io.Reader, jsonVar interface{}) error
	GetAllProjects(ctx context.Context, waitChan chan<- error) error
	UpdateAllProjects(ctx context.Context) error
	SetToken(token string)
	SetUrl(url string)
	GetUrl() string
	Run(ctx context.Context) error
	GetName() string
}

type RepositoryInterface interface {
	CompareBranches(source, target string) (bool, error)
	GetAllMRs(ctx context.Context) error
	GetAllBranches(ctx context.Context) error
	Run(ctx context.Context) error
}
