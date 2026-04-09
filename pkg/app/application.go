package app

import (
	"context"

	"github.com/pkg/errors"
)

var ErrNotImplemented = errors.New("not implemented")

type Application struct{}

func New() *Application {
	return &Application{}
}

type CompileSummary struct {
	File      string
	SessionID string
}

func (a *Application) DiscoveryList(ctx context.Context, kind string) ([]map[string]any, error) {
	_ = ctx
	_ = kind
	return nil, ErrNotImplemented
}

func (a *Application) CompileFile(ctx context.Context, file string) (*CompileSummary, error) {
	_ = ctx
	_ = file
	return nil, ErrNotImplemented
}

func (a *Application) RecordFile(ctx context.Context, file string) error {
	_ = ctx
	_ = file
	return ErrNotImplemented
}
