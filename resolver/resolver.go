package resolver

import (
	"context"

	dto "github.com/prometheus/client_model/go"
)

// Result is the struct that contains results from Resolver
type Result struct {
	IP    string
	Label *dto.LabelPair
}

// Resolver is the base interface for any host resolver
type Resolver interface {
	Resolve(ctx context.Context, target string) ([]Result, error)
}
